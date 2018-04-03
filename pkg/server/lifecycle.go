package server

import (
	"github.com/drausin/libri/libri/common/errors"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage/postgres/migrations"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	bindata "github.com/mattes/migrate/source/go-bindata"
	"google.golang.org/grpc"
)

// Start starts the server and eviction routines.
func Start(config *Config, up chan *Catalog) error {
	c, err := newCatalog(config)
	if err != nil {
		return err
	}

	if err := c.maybeMigrateDB(); err != nil {
		return err
	}

	registerServer := func(s *grpc.Server) { api.RegisterCatalogServer(s, c) }
	return c.Serve(registerServer, func() { up <- c })
}

// StopServer handles cleanup involved in closing down the server.
func (x *Catalog) StopServer() {
	x.BaseServer.StopServer()
	err := x.storer.Close()
	errors.MaybePanic(err)
}

func (x *Catalog) maybeMigrateDB() error {
	if x.config.Storage.Type != bstorage.Postgres {
		return nil
	}

	m := bstorage.NewBindataMigrator(
		x.config.DBUrl,
		bindata.Resource(migrations.AssetNames(), migrations.Asset),
		&bstorage.ZapLogger{Logger: x.Logger},
	)
	return m.Up()
}
