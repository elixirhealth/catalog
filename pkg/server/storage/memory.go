package storage

import (
	"bytes"
	"container/heap"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/drausin/libri/libri/common/id"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"go.uber.org/zap"
)

// NewMemory creates a new Storer backed by an in-memory map with the given parameters and logger.
func NewMemory(params *Parameters, logger *zap.Logger) Storer {
	return &memoryStorer{
		storedPRs: make(map[string]*PublicationReceipt),
		params:    params,
		logger:    logger,
	}
}

type memoryStorer struct {
	storedPRs map[string]*PublicationReceipt
	params    *Parameters
	logger    *zap.Logger
	mu        sync.Mutex
}

func (f *memoryStorer) Put(pr *api.PublicationReceipt) error {
	if err := api.ValidatePublicationReceipt(pr); err != nil {
		return err
	}
	envKeyHex := id.Hex(pr.EnvelopeKey)
	spr := encodeStoredPubReceipt(pr)
	spr.EnvelopeKey = datastore.NameKey(publicationReceiptKind, envKeyHex, nil)
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, in := f.storedPRs[envKeyHex]; in {
		f.logger.Debug("publication receipt already exists",
			zap.String(logEnvelopeKey, envKeyHex))
		return nil
	}
	f.storedPRs[envKeyHex] = spr
	f.logger.Debug("stored new publication", zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
	return nil
}

func (f *memoryStorer) Search(
	filters *SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	if err := validateSearchFilters(filters); err != nil {
		return nil, err
	}
	if limit > MaxSearchLimit {
		return nil, ErrSearchLimitTooLarge
	}
	storedResults := &publicationReceipts{}
	f.mu.Lock()
	for _, spr := range f.storedPRs {
		pr, err := decodeStoredPubReceipt(spr)
		if err != nil {
			return nil, err
		}
		if matchesFilter(filters, pr) {
			heap.Push(storedResults, spr)
			if storedResults.Len() > int(limit) {
				heap.Pop(storedResults)
			}
		}
	}
	f.mu.Unlock()
	return storedResultsToList(storedResults)
}

func matchesFilter(filters *SearchFilters, pr *api.PublicationReceipt) bool {
	// assume that at least one filter is set per validateSearchFilters
	if filters.EntryKey != nil && !bytes.Equal(filters.EntryKey, pr.EntryKey) {
		return false
	}
	if filters.AuthorPublicKey != nil &&
		!bytes.Equal(filters.AuthorPublicKey, pr.AuthorPublicKey) {
		return false
	}
	if filters.ReaderPublicKey != nil &&
		!bytes.Equal(filters.ReaderPublicKey, pr.ReaderPublicKey) {
		return false
	}
	if filters.Before != 0 && pr.ReceivedTime >= filters.Before {
		return false
	}
	return true
}
