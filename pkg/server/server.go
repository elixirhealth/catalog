package server

import (
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/catalog/pkg/server/storage"
	"github.com/elxirhealth/service-base/pkg/server"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Catalog implements the CatalogServer interface.
type Catalog struct {
	*server.BaseServer
	config *Config

	storer storage.Storer
}

// newCatalog creates a new CatalogServer from the given config.
func newCatalog(config *Config) (*Catalog, error) {
	baseServer := server.NewBaseServer(config.BaseConfig)
	storer, err := getStorer(config, baseServer.Logger)
	if err != nil {
		return nil, err
	}
	return &Catalog{
		BaseServer: baseServer,
		config:     config,
		storer:     storer,
	}, nil
}

// Put adds a publication to the catalog.
func (x *Catalog) Put(ctx context.Context, rq *api.PutRequest) (*api.PutResponse, error) {
	x.Logger.Debug("received Put request", logPutRequestFields(rq)...)
	if err := api.ValidatePutRequest(rq); err != nil {
		return nil, err
	}
	if err := x.storer.Put(rq.Value); err != nil {
		return nil, err
	}
	x.Logger.Info("put publication receipt", logPutRequestFields(rq)...)
	return &api.PutResponse{}, nil
}

// Search finds publications in the catalog matching the given filter criteria.
func (x *Catalog) Search(ctx context.Context, rq *api.SearchRequest) (*api.SearchResponse, error) {
	x.Logger.Debug("received search request", logSearchRequestFields(rq)...)
	filters := &storage.SearchFilters{
		EntryKey:        rq.EntryKey,
		AuthorPublicKey: rq.AuthorPublicKey,
		ReaderPublicKey: rq.ReaderPublicKey,
		Before:          rq.Before,
	}
	result, err := x.storer.Search(filters, uint(rq.Limit))
	if err != nil {
		return nil, err
	}
	x.Logger.Info("found search results", zap.Int(logNResults, len(result)))
	return &api.SearchResponse{Result: result}, nil
}
