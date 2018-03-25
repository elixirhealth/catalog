package server

import (
	"testing"

	"github.com/elixirhealth/catalog/pkg/server/storage"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	c := NewDefaultConfig()
	assert.NotNil(t, c)
	assert.NotEmpty(t, c.Storage)
}

func TestConfig_WithStorage(t *testing.T) {
	c1, c2, c3 := &Config{}, &Config{}, &Config{}
	c1.WithDefaultStorage()
	assert.Equal(t, c1.Storage.Type, c2.WithStorage(nil).Storage.Type)
	assert.NotEqual(t,
		c1.Storage.Type,
		c3.WithStorage(
			&storage.Parameters{Type: bstorage.DataStore},
		).Storage.Type,
	)
}

func TestConfig_WithGCPProjectID(t *testing.T) {
	c1 := &Config{}
	p := "project-ID"
	c1.WithGCPProjectID(p)
	assert.Equal(t, p, c1.GCPProjectID)
}
