package server

import (
	"testing"

	"github.com/elxirhealth/catalog/pkg/server/storage"
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
	assert.Equal(t, c1.Storage.StorageType, c2.WithStorage(nil).Storage.StorageType)
	assert.NotEqual(t,
		c1.Storage.StorageType,
		c3.WithStorage(
			&storage.Parameters{StorageType: storage.DataStore},
		).Storage.StorageType,
	)
}

func TestConfig_WithGCPProjectID(t *testing.T) {
	c1 := &Config{}
	p := "project-ID"
	c1.WithGCPProjectID(p)
	assert.Equal(t, p, c1.GCPProjectID)
}
