package server

import (
	"errors"

	"github.com/elxirhealth/catalog/pkg/server/storage"
	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"go.uber.org/zap"
)

var (

	// ErrInvalidStorageType indicates when a storage type is not expected.
	ErrInvalidStorageType = errors.New("invalid storage type")
)

func getStorer(config *Config, logger *zap.Logger) (storage.Storer, error) {
	switch config.Storage.Type {
	case bstorage.DataStore:
		return storage.NewDatastore(config.GCPProjectID, config.Storage, logger)
	case bstorage.Memory:
		return storage.NewMemory(config.Storage, logger), nil
	default:
		return nil, ErrInvalidStorageType
	}
}
