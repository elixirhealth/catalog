package server

import (
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"google.golang.org/grpc"
)

// Start starts the server and eviction routines.
func Start(config *Config, up chan *Catalog) error {
	c, err := newCatalog(config)
	if err != nil {
		return err
	}

	registerServer := func(s *grpc.Server) { api.RegisterCatalogServer(s, c) }
	return c.Serve(registerServer, func() { up <- c })
}
