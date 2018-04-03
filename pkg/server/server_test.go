package server

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/id"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/server"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestNewCatalog_ok(t *testing.T) {
	config := NewDefaultConfig()
	c, err := newCatalog(config)
	assert.Nil(t, err)
	assert.Equal(t, config, c.config)
	assert.NotEmpty(t, c.storer)
}

func TestNewCatalog_err(t *testing.T) {
	badConfigs := map[string]*Config{
		"empty ProjectID": NewDefaultConfig().WithStorage(
			&storage.Parameters{Type: bstorage.DataStore},
		),
	}
	for desc, badConfig := range badConfigs {
		c, err := newCatalog(badConfig)
		assert.NotNil(t, err, desc)
		assert.Nil(t, c)
	}
}

func TestCatalog_Put_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	x := &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer:     &fixedStorer{},
	}
	rq := &api.PutRequest{
		Value: &api.PublicationReceipt{
			EnvelopeKey:     util.RandBytes(rng, id.Length),
			EntryKey:        util.RandBytes(rng, id.Length),
			AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
			ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
			ReceivedTime:    api.ToEpochMicros(time.Now()),
		},
	}

	rp, err := x.Put(context.Background(), rq)
	assert.Nil(t, err)
	assert.NotNil(t, rp)
}

func TestCatalog_Put_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	// bad request
	x := &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer:     &fixedStorer{},
	}
	rq := &api.PutRequest{
		Value: &api.PublicationReceipt{
			// missing EnvelopeKey
			EntryKey:        util.RandBytes(rng, id.Length),
			AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
			ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
			ReceivedTime:    api.ToEpochMicros(time.Now()),
		},
	}

	rp, err := x.Put(context.Background(), rq)
	assert.NotNil(t, err)
	assert.Nil(t, rp)

	// put error
	x = &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer:     &fixedStorer{putErr: errors.New("some Put error")},
	}
	rq = &api.PutRequest{
		Value: testPubReceipt(rng),
	}

	rp, err = x.Put(context.Background(), rq)
	assert.NotNil(t, err)
	assert.Nil(t, rp)
}

func TestCatalog_Search_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	x := &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer: &fixedStorer{
			searchResult: []*api.PublicationReceipt{
				testPubReceipt(rng),
			},
		},
	}
	rq := &api.SearchRequest{
		EntryKey: util.RandBytes(rng, id.Length),
		Limit:    storage.MaxSearchLimit,
	}

	rp, err := x.Search(context.Background(), rq)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rp.Result))
}

func TestCatalog_Search_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	x := &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer:     &fixedStorer{},
	}
	rq := &api.SearchRequest{
		EntryKey: util.RandBytes(rng, id.Length),
	}

	rp, err := x.Search(context.Background(), rq)
	assert.NotNil(t, err)
	assert.Nil(t, rp)

	x = &Catalog{
		BaseServer: server.NewBaseServer(server.NewDefaultBaseConfig()),
		storer: &fixedStorer{
			searchErr: errors.New("some Search error"),
		},
	}
	rq = &api.SearchRequest{
		EntryKey: util.RandBytes(rng, id.Length),
		Limit:    storage.MaxSearchLimit,
	}

	rp, err = x.Search(context.Background(), rq)
	assert.NotNil(t, err)
	assert.Nil(t, rp)
}

func testPubReceipt(rng *rand.Rand) *api.PublicationReceipt {
	return &api.PublicationReceipt{
		EnvelopeKey:     util.RandBytes(rng, id.Length),
		EntryKey:        util.RandBytes(rng, id.Length),
		AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReceivedTime:    api.ToEpochMicros(time.Now()),
	}
}

type fixedStorer struct {
	putErr       error
	searchResult []*api.PublicationReceipt
	searchErr    error
}

func (f *fixedStorer) Put(pub *api.PublicationReceipt) error {
	return f.putErr
}

func (f *fixedStorer) Search(
	filters *storage.SearchFilters, limit uint32,
) ([]*api.PublicationReceipt, error) {
	return f.searchResult, f.searchErr
}

func (f *fixedStorer) Close() error {
	return nil
}
