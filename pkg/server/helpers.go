package server

import (
	"errors"

	"github.com/elxirhealth/catalog/pkg/server/storage"
	"go.uber.org/zap"
)

var (

	// ErrInvalidStorageType indicates when a storage type is not expected.
	ErrInvalidStorageType = errors.New("invalid storage type")
)

func getStorer(config *Config, logger *zap.Logger) (storage.Storer, error) {
	switch config.Storage.StorageType {
	case storage.DataStore:
		return storage.NewDatastore(config.GCPProjectID, config.Storage, logger)
	case storage.InMemory:
		panic("not implemented yet")
	default:
		return nil, ErrInvalidStorageType
	}
}
