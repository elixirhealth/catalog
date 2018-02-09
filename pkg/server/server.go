package server

import (
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/catalog/pkg/server/storage"
	"github.com/elxirhealth/service-base/pkg/server"
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

	// TODO add other init

	return &Catalog{
		BaseServer: baseServer,
		config:     config,
		// TODO add other things
	}, nil
}

// Put adds a publication to the catalog.
func (x *Catalog) Put(ctx context.Context, rq *api.PutRequest) (*api.PutResponse, error) {
	// TODO (drausin) debug log request
	// TODO (drausin) validate rq
	if err := x.storer.Put(rq.Pub); err != nil {
		return nil, err
	}
	// TODO (drausin) info log success
	return &api.PutResponse{}, nil
}

// Search finds publications in the catalog matching the given filter criteria.
func (x *Catalog) Search(ctx context.Context, rq *api.SearchRequest) (*api.SearchResponse, error) {
	// TODO (drausin) debug log request
	// TODO (drausin) validate rq
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
	// TODO (drausin) info log success
	return &api.SearchResponse{Result: result}, nil
}
