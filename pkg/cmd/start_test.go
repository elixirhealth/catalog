package cmd

import (
	"testing"
	"time"

	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
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
	dbURL := "some URL"
	searchTimeout := 10 * time.Second
	storageInMemory := false
	storagePostgres := true

	viper.Set(serverPortFlag, serverPort)
	viper.Set(metricsPortFlag, metricsPort)
	viper.Set(profilerPortFlag, profilerPort)
	viper.Set(logLevelFlag, logLevel)
	viper.Set(profileFlag, profile)
	viper.Set(dbURLFlag, dbURL)
	viper.Set(searchTimeoutFlag, searchTimeout)
	viper.Set(storageMemoryFlag, storageInMemory)
	viper.Set(storagePostgresFlag, storagePostgres)

	c, err := getCatalogConfig()
	assert.Nil(t, err)
	assert.Equal(t, serverPort, c.ServerPort)
	assert.Equal(t, metricsPort, c.MetricsPort)
	assert.Equal(t, profilerPort, c.ProfilerPort)
	assert.Equal(t, logLevel, c.LogLevel.String())
	assert.Equal(t, profile, c.Profile)
	assert.Equal(t, searchTimeout, c.Storage.SearchTimeout)
	assert.NotEmpty(t, c.Storage.GetTimeout)
	assert.NotEmpty(t, c.Storage.PutTimeout)
	assert.Equal(t, dbURL, c.DBUrl)
	assert.Equal(t, bstorage.Postgres, c.Storage.Type)
}

func TestGetDBUrl(t *testing.T) {

	// no DB password set should just return URL
	unsafeDbURL := "postgres://admin@127.0.0.1:5432/catalog?sslmode=disable&password=changeme"
	viper.Set(dbURLFlag, unsafeDbURL)
	assert.Equal(t, unsafeDbURL, getDBUrl())

	// set DB password should be append to end
	safeDbURL := "postgres://admin@127.0.0.1:5432/catalog?sslmode=disable"
	dbPass := "changeme"
	viper.Set(dbURLFlag, safeDbURL)
	viper.Set(dbPasswordFlag, dbPass)
	assert.Equal(t, unsafeDbURL, getDBUrl())
}

func TestGetCacheStorageType(t *testing.T) {
	viper.Set(storageMemoryFlag, true)
	viper.Set(storagePostgresFlag, false)
	st, err := getStorageType()
	assert.Nil(t, err)
	assert.Equal(t, bstorage.Memory, st)

	viper.Set(storageMemoryFlag, false)
	viper.Set(storagePostgresFlag, true)
	st, err = getStorageType()
	assert.Nil(t, err)
	assert.Equal(t, bstorage.Postgres, st)

	viper.Set(storageMemoryFlag, true)
	viper.Set(storagePostgresFlag, true)
	st, err = getStorageType()
	assert.Equal(t, errMultipleStorageTypes, err)
	assert.Equal(t, bstorage.Unspecified, st)

	viper.Set(storageMemoryFlag, false)
	viper.Set(storagePostgresFlag, false)
	st, err = getStorageType()
	assert.Equal(t, errNoStorageType, err)
	assert.Equal(t, bstorage.Unspecified, st)
}
