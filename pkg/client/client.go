package client

import (
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"google.golang.org/grpc"
)

// NewInsecure returns a new CatalogClient without any TLS on the connection.
func NewInsecure(address string) (api.CatalogClient, error) {
	cc, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return api.NewCatalogClient(cc), nil
}
