package server

import (
	"errors"

	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/catalog/pkg/server/storage/datastore"
	"github.com/elixirhealth/catalog/pkg/server/storage/memory"
	"github.com/elixirhealth/catalog/pkg/server/storage/postgres"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap"
)

var (

	// ErrInvalidStorageType indicates when a storage type is not expected.
	ErrInvalidStorageType = errors.New("invalid storage type")
)

func getStorer(config *Config, logger *zap.Logger) (storage.Storer, error) {
	switch config.Storage.Type {
	case bstorage.Postgres:
		return postgres.New(config.DBUrl, config.Storage, logger)
	case bstorage.DataStore:
		return datastore.New(config.GCPProjectID, config.Storage, logger)
	case bstorage.Memory:
		return memory.New(config.Storage, logger), nil
	default:
		return nil, ErrInvalidStorageType
	}
}
