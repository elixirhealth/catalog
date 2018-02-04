package server

import (
	"github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/service-base/pkg/server"
	"golang.org/x/net/context"
)

// Catalog implements the CatalogServer interface.
type Catalog struct {
	*server.BaseServer
	config *Config

	// TODO add other things here
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

func (x *Catalog) Put(context.Context, *catalogapi.PutRequest) (*catalogapi.PutResponse, error) {
	panic("implement me")
}

func (x *Catalog) Search(
	context.Context, *catalogapi.SearchRequest,
) (*catalogapi.SearchResponse, error) {
	panic("implement me")
}
