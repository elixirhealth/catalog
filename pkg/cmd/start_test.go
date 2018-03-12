package cmd

import (
	"testing"
	"time"

	bstorage "github.com/elxirhealth/service-base/pkg/server/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestGetCatalogConfig(t *testing.T) {
	serverPort := uint(1234)
	metricsPort := uint(5678)
	profilerPort := uint(9012)
	logLevel := zapcore.DebugLevel.String()
	profile := true
	gcpProjectID := "some project"
	searchTimeout := 10 * time.Second
	storageInMemory := false
	storageDataStore := true

	viper.Set(serverPortFlag, serverPort)
	viper.Set(metricsPortFlag, metricsPort)
	viper.Set(profilerPortFlag, profilerPort)
	viper.Set(logLevelFlag, logLevel)
	viper.Set(profileFlag, profile)
	viper.Set(gcpProjectIDFlag, gcpProjectID)
	viper.Set(searchTimeoutFlag, searchTimeout)
	viper.Set(storageMemoryFlag, storageInMemory)
	viper.Set(storageDataStoreFlag, storageDataStore)

	c, err := getCatalogConfig()
	assert.Nil(t, err)
	assert.Equal(t, serverPort, c.ServerPort)
	assert.Equal(t, metricsPort, c.MetricsPort)
	assert.Equal(t, profilerPort, c.ProfilerPort)
	assert.Equal(t, logLevel, c.LogLevel.String())
	assert.Equal(t, profile, c.Profile)
	assert.Equal(t, gcpProjectID, c.GCPProjectID)
	assert.Equal(t, searchTimeout, c.Storage.SearchQueryTimeout)
	assert.Equal(t, bstorage.DataStore, c.Storage.Type)
}

func TestGetCacheStorageType(t *testing.T) {
	viper.Set(storageMemoryFlag, true)
	viper.Set(storageDataStoreFlag, false)
	st, err := getStorageType()
	assert.Nil(t, err)
	assert.Equal(t, bstorage.Memory, st)

	viper.Set(storageMemoryFlag, false)
	viper.Set(storageDataStoreFlag, true)
	st, err = getStorageType()
	assert.Nil(t, err)
	assert.Equal(t, bstorage.DataStore, st)

	viper.Set(storageMemoryFlag, true)
	viper.Set(storageDataStoreFlag, true)
	st, err = getStorageType()
	assert.Equal(t, errMultipleStorageTypes, err)
	assert.Equal(t, bstorage.Unspecified, st)

	viper.Set(storageMemoryFlag, false)
	viper.Set(storageDataStoreFlag, false)
	st, err = getStorageType()
	assert.Equal(t, errNoStorageType, err)
	assert.Equal(t, bstorage.Unspecified, st)
}
