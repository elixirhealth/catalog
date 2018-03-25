package cmd

import (
	"fmt"
	"sync"
	"testing"

	"github.com/elixirhealth/catalog/pkg/server"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestTestIO(t *testing.T) {
	nPuts := uint(32)
	nSearches := uint(4)

	// start in-memory catalog
	config := server.NewDefaultConfig()
	config.Storage.Type = bstorage.Memory
	config.LogLevel = zapcore.DebugLevel
	config.ServerPort = 10200
	config.MetricsPort = 10201

	up := make(chan *server.Catalog, 1)
	wg1 := new(sync.WaitGroup)
	wg1.Add(1)
	go func(wg2 *sync.WaitGroup) {
		defer wg2.Done()
		err := server.Start(config, up)
		assert.Nil(t, err)
	}(wg1)

	c := <-up
	viper.Set(catalogsFlag, fmt.Sprintf("localhost:%d", config.ServerPort))
	viper.Set(nPutsFlag, nPuts)
	viper.Set(nSearchesFlag, nSearches)

	err := testIO()
	assert.Nil(t, err)

	c.StopServer()
	wg1.Wait()
}
