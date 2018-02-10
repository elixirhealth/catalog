package storage

import (
	"container/heap"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/drausin/libri/libri/common/id"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const (
	// MaxSearchLimit is the maximum number of search results that can be returned in a single
	// search query.
	MaxSearchLimit = uint32(128)

	publicationReceiptKind = "publicationReceipt"
	secsPerDay             = 60 * 60 * 24
)

var (
	// ErrSearchLimitTooLarge denotes when the search limit given is larger than MaxSearchLimit.
	ErrSearchLimitTooLarge = fmt.Errorf("search limit larger than maximum value %d",
		MaxSearchLimit)
)

// PublicationReceipt represents the publication and its (usually) earliest receipt of a particular
// envelope.
type PublicationReceipt struct {
	EnvelopeKey     *datastore.Key `datastore:"__key__"`
	EntryKey        string         `datastore:"entry_key"`
	AuthorPublicKey string         `datastore:"author_public_key"`
	ReaderPublicKey string         `datastore:"reader_public_key"`
	ReceivedDate    int32          `datastore:"received_date"`
	ReceivedTime    time.Time      `datastore:"received_time,noindex"`
}

type datastoreStorer struct {
	client storage.DatastoreClient
	iter   storage.DatastoreIterator
	params *Parameters
	logger *zap.Logger
}

// NewDatastore creates a new Storer backed by a DataStore instance in the given GCP project with
// the given parameters and logger.
func NewDatastore(gcpProjectID string, params *Parameters, logger *zap.Logger) (Storer, error) {
	client, err := datastore.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		return nil, err
	}
	return &datastoreStorer{
		client: &storage.DatastoreClientImpl{Inner: client},
		iter:   &storage.DatastoreIteratorImpl{},
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
	err := d.client.Get(pubKey, &PublicationReceipt{})
	if err != nil && err != datastore.ErrNoSuchEntity {
		return err
	} else if err == nil {
		// nothing to do if already exists
		d.logger.Debug("publication receipt already exists",
			zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
		return nil
	}
	spr := encodeStoredPubReceipt(pr)
	if _, err = d.client.Put(pubKey, spr); err != nil {
		return err
	}
	d.logger.Debug("stored new publication", zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
	return nil
}

// Search searches for publications matching the given filters.
func (d *datastoreStorer) Search(
	fs *SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	if err := validateSearchFilters(fs); err != nil {
		return nil, err
	}
	if limit > MaxSearchLimit {
		return nil, ErrSearchLimitTooLarge
	}
	storedResults := &publicationReceipts{}
	ctx, cancel := context.WithTimeout(context.Background(), d.params.SearchQueryTimeout)
	q := getSearchQuery(fs)
	iter := d.client.Run(ctx, q)
	d.iter.Init(iter)
	for {
		stored := &PublicationReceipt{}
		if _, err := d.iter.Next(stored); err == iterator.Done {
			// no more results
			break
		} else if err != nil {
			cancel()
			return nil, err
		}
		d.logger.Debug("processing search result",
			zap.String(logEnvelopeKey, stored.EnvelopeKey.Name))
		if done := processStored(stored, fs.Before, storedResults, limit); done {
			break
		}
	}
	cancel()

	result := make([]*api.PublicationReceipt, storedResults.Len())
	for i := storedResults.Len() - 1; i >= 0; i-- {
		storedResult := heap.Pop(storedResults).(*PublicationReceipt)
		v, err := decodeStoredPubReceipt(storedResult)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

func processStored(
	stored *PublicationReceipt,
	beforeFilter int64,
	storedResults *publicationReceipts,
	limit uint32,
) bool {
	before := stored.ReceivedTime.Before(api.FromEpochMicros(beforeFilter))
	if beforeFilter != 0 && !before {
		// don't add a result if we have a before filter and the result is after it
		return false
	}
	dayChange := storedResults.Len() > 0 &&
		stored.ReceivedDate < storedResults.Peak().ReceivedDate
	if storedResults.Len() == int(limit) && dayChange {
		// if we have {{limit}} results and the current received date is
		// earlier/less than the latest, we know we've found the latest {{limit}}
		// results
		return true
	}
	heap.Push(storedResults, stored)
	if storedResults.Len() > int(limit) {
		heap.Pop(storedResults)
	}
	return false
}

func getSearchQuery(f *SearchFilters) *datastore.Query {
	q := datastore.NewQuery(publicationReceiptKind).Order("-received_date")
	if f.AuthorPublicKey != nil {
		q = q.Filter("author_public_key = ", hex.EncodeToString(f.AuthorPublicKey))
	}
	if f.ReaderPublicKey != nil {
		q = q.Filter("reader_public_key = ", hex.EncodeToString(f.ReaderPublicKey))
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
		ReaderPublicKey: hex.EncodeToString(pr.ReaderPublicKey),
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
	pr.AuthorPublicKey = v
	v, err = hex.DecodeString(s.ReaderPublicKey)
	if err != nil {
		return nil, err
	}
	pr.ReaderPublicKey = v

	return pr, nil
}
