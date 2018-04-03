package memory

import (
	"bytes"
	"container/heap"
	"sync"

	"github.com/drausin/libri/libri/common/id"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"go.uber.org/zap"
)

// New creates a new Storer backed by an in-memory map with the given parameters and logger.
func New(params *storage.Parameters, logger *zap.Logger) storage.Storer {
	return &memoryStorer{
		prs:    make(map[string]*api.PublicationReceipt),
		params: params,
		logger: logger,
	}
}

type memoryStorer struct {
	prs    map[string]*api.PublicationReceipt
	params *storage.Parameters
	logger *zap.Logger
	mu     sync.Mutex
}

func (f *memoryStorer) Put(pr *api.PublicationReceipt) error {
	if err := api.ValidatePublicationReceipt(pr); err != nil {
		return err
	}
	envKeyHex := id.Hex(pr.EnvelopeKey)
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, in := f.prs[envKeyHex]; in {
		f.logger.Debug("publication receipt already exists",
			zap.String(logEnvelopeKey, envKeyHex))
		return nil
	}
	f.prs[envKeyHex] = pr
	f.logger.Debug("stored new publication", zap.String(logEnvelopeKey, id.Hex(pr.EnvelopeKey)))
	return nil
}

func (f *memoryStorer) Search(
	filters *storage.SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	if err := storage.ValidateSearchFilters(filters); err != nil {
		return nil, err
	}
	if limit > storage.MaxSearchLimit {
		return nil, storage.ErrSearchLimitTooLarge
	}
	results := &storage.PublicationReceipts{}
	f.mu.Lock()
	for _, pr := range f.prs {
		if matchesFilter(filters, pr) {
			heap.Push(results, pr)
			if results.Len() > int(limit) {
				heap.Pop(results)
			}
		}
	}
	f.mu.Unlock()
	return results.PopList(), nil
}

func (f *memoryStorer) Close() error {
	return nil
}

func matchesFilter(filters *storage.SearchFilters, pr *api.PublicationReceipt) bool {
	// assume that at least one filter is set per ValidateSearchFilters
	if bytesNotMatch(filters.EntryKey, pr.EntryKey) {
		return false
	}
	if bytesNotMatch(filters.AuthorPublicKey, pr.AuthorPublicKey) {
		return false
	}
	if stringsNotMatch(filters.AuthorEntityID, pr.AuthorEntityId) {
		return false
	}
	if bytesNotMatch(filters.ReaderPublicKey, pr.ReaderPublicKey) {
		return false
	}
	if stringsNotMatch(filters.ReaderEntityID, pr.ReaderEntityId) {
		return false
	}
	if filters.Before != 0 && pr.ReceivedTime >= filters.Before {
		return false
	}
	if filters.After != 0 && pr.ReceivedTime < filters.After {
		return false
	}
	return true
}

func stringsNotMatch(filter, value string) bool {
	return filter != "" && filter != value
}

func bytesNotMatch(filter, value []byte) bool {
	return filter != nil && !bytes.Equal(filter, value)
}
