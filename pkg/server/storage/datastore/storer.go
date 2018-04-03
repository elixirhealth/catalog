package datastore

import (
	"container/heap"
	"context"
	"encoding/hex"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/drausin/libri/libri/common/id"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const (
	publicationReceiptKind = "publicationReceipt"
	secsPerDay             = 60 * 60 * 24
)

// PublicationReceipt represents the publication and its (usually) earliest receipt of a particular
// envelope.
type PublicationReceipt struct {
	EnvelopeKey     *datastore.Key `datastore:"__key__"`
	EntryKey        string         `datastore:"entry_key"`
	AuthorPublicKey string         `datastore:"author_public_key"`
	AuthorEntityID  string         `datastore:"author_entity_id"`
	ReaderPublicKey string         `datastore:"reader_public_key"`
	ReaderEntityID  string         `datastore:"reader_entity_id"`
	ReceivedDate    int32          `datastore:"received_date"`
	ReceivedTime    time.Time      `datastore:"received_time,noindex"`
}

type datastoreStorer struct {
	client bstorage.DatastoreClient
	iter   bstorage.DatastoreIterator
	params *storage.Parameters
	logger *zap.Logger
}

// New creates a new Storer backed by a DataStore instance in the given GCP project with
// the given parameters and logger.
func New(
	gcpProjectID string, params *storage.Parameters, logger *zap.Logger,
) (storage.Storer, error) {
	client, err := datastore.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		return nil, err
	}
	return &datastoreStorer{
		client: &bstorage.DatastoreClientImpl{Inner: client},
		iter:   &bstorage.DatastoreIteratorImpl{},
		params: params,
		logger: logger,
	}, nil
}

// Put stores a new publication if one doesn't already exist for the given envelope key.
func (d *datastoreStorer) Put(pr *api.PublicationReceipt) error {
	if err := api.ValidatePublicationReceipt(pr); err != nil {
		return err
	}
	pubKey := datastore.NameKey(publicationReceiptKind, id.Hex(pr.EnvelopeKey), nil)
	ctx, cancel := context.WithTimeout(context.Background(), d.params.GetTimeout)
	err := d.client.Get(ctx, pubKey, &PublicationReceipt{})
	cancel()
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	} else if err == nil {
		// nothing to do if already exists
		d.logger.Debug("publication receipt already exists",
			zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
		return nil
	}
	spr := encodeStoredPubReceipt(pr)
	ctx, cancel = context.WithTimeout(context.Background(), d.params.PutTimeout)
	defer cancel()
	if _, err = d.client.Put(ctx, pubKey, spr); err != nil {
		return err
	}
	d.logger.Debug("stored new publication", zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
	return nil
}

// Search searches for publications matching the given filters.
func (d *datastoreStorer) Search(
	fs *storage.SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	if err := storage.ValidateSearchFilters(fs); err != nil {
		return nil, err
	}
	if limit > storage.MaxSearchLimit {
		return nil, storage.ErrSearchLimitTooLarge
	}
	results := &storage.PublicationReceipts{}
	ctx, cancel := context.WithTimeout(context.Background(), d.params.SearchTimeout)
	defer cancel()
	q := getSearchQuery(fs)
	iter := d.client.Run(ctx, q)
	d.iter.Init(iter)
	for {
		stored := &PublicationReceipt{}
		if _, err := d.iter.Next(stored); err == iterator.Done {
			// no more results
			break
		} else if err != nil {
			return nil, err
		}
		d.logger.Debug("processing search result",
			zap.String(logEnvelopeKey, stored.EnvelopeKey.Name))
		done, err := processStored(stored, fs.Before, fs.After, results, limit)
		if err != nil {
			return nil, err
		}
		if done {
			break
		}
	}
	return results.PopList(), nil
}

func (f *datastoreStorer) Close() error {
	return nil
}

func processStored(
	stored *PublicationReceipt,
	beforeFilter int64,
	afterFilter int64,
	results *storage.PublicationReceipts,
	limit uint32,
) (bool, error) {
	before := stored.ReceivedTime.Before(api.FromEpochMicros(beforeFilter))
	after := !stored.ReceivedTime.Before(api.FromEpochMicros(afterFilter)) // inclusive after
	if beforeFilter != 0 && !before {
		// don't add a result if we have a before filter and the result is after it
		return false, nil
	}
	if afterFilter != 0 && !after {
		// don't add a result if we have an after filter and the result is before it
		return false, nil
	}
	dayChange := results.Len() > 0 &&
		stored.ReceivedDate < int32(results.Peak().ReceivedTime/secsPerDay)
	if results.Len() == int(limit) && dayChange {
		// if we have {{limit}} results and the current received date is
		// earlier/less than the latest, we know we've found the latest {{limit}}
		// results
		return true, nil
	}
	pr, err := decodeStoredPubReceipt(stored)
	if err != nil {
		return false, err
	}
	heap.Push(results, pr)
	if results.Len() > int(limit) {
		heap.Pop(results)
	}
	return false, nil
}

func getSearchQuery(f *storage.SearchFilters) *datastore.Query {
	q := datastore.NewQuery(publicationReceiptKind).Order("-received_date")
	if f.AuthorPublicKey != nil {
		q = q.Filter("author_public_key = ", hex.EncodeToString(f.AuthorPublicKey))
	}
	if f.AuthorEntityID != "" {
		q = q.Filter("author_entity_id = ", f.AuthorEntityID)
	}
	if f.ReaderPublicKey != nil {
		q = q.Filter("reader_public_key = ", hex.EncodeToString(f.ReaderPublicKey))
	}
	if f.ReaderEntityID != "" {
		q = q.Filter("reader_entity_id = ", f.ReaderEntityID)
	}
	if f.EntryKey != nil {
		q = q.Filter("entry_key = ", hex.EncodeToString(f.EntryKey))
	}
	if f.Before != 0 {
		// include current day of f.Before, so "before" date becomes day after f.Before date
		beforeDate := int32(f.Before/secsPerDay) + 1
		q = q.Filter("received_date < ", beforeDate)
	}
	return q
}

func encodeStoredPubReceipt(pr *api.PublicationReceipt) *PublicationReceipt {
	return &PublicationReceipt{
		EntryKey:        hex.EncodeToString(pr.EntryKey),
		AuthorPublicKey: hex.EncodeToString(pr.AuthorPublicKey),
		AuthorEntityID:  pr.AuthorEntityId,
		ReaderPublicKey: hex.EncodeToString(pr.ReaderPublicKey),
		ReaderEntityID:  pr.ReaderEntityId,
		ReceivedDate:    int32(pr.ReceivedTime / secsPerDay),
		ReceivedTime:    api.FromEpochMicros(pr.ReceivedTime),
	}
}

func decodeStoredPubReceipt(s *PublicationReceipt) (*api.PublicationReceipt, error) {
	pr := &api.PublicationReceipt{
		ReceivedTime: api.ToEpochMicros(s.ReceivedTime),
	}
	v, err := hex.DecodeString(s.EnvelopeKey.Name)
	if err != nil {
		return nil, err
	}
	pr.EnvelopeKey = v
	v, err = hex.DecodeString(s.EntryKey)
	if err != nil {
		return nil, err
	}
	pr.EntryKey = v
	v, err = hex.DecodeString(s.AuthorPublicKey)
	if err != nil {
		return nil, err
	}
	pr.AuthorEntityId = s.AuthorEntityID
	pr.AuthorPublicKey = v
	v, err = hex.DecodeString(s.ReaderPublicKey)
	if err != nil {
		return nil, err
	}
	pr.ReaderPublicKey = v
	pr.ReaderEntityId = s.ReaderEntityID

	return pr, nil
}
