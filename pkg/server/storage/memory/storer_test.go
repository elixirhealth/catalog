package memory

import (
	"math/rand"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemoryStorer_PutSearch(t *testing.T) {
	lg := logging.NewDevInfoLogger()
	s := New(storage.NewDefaultParameters(), lg)
	storage.TestStorerPutSearch(t, s)
}

func TestMemoryStorer_Put_ok(t *testing.T) {
	lg, params := zap.NewNop(), storage.NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))
	pr := &api.PublicationReceipt{
		EnvelopeKey:     util.RandBytes(rng, id.Length),
		EntryKey:        util.RandBytes(rng, id.Length),
		AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReceivedTime:    api.ToEpochMicros(time.Now()),
	}

	// mock PR already existing
	d := &memoryStorer{
		prs: map[string]*api.PublicationReceipt{
			id.Hex(pr.EnvelopeKey): pr,
		},
		logger: lg,
		params: params,
	}
	err := d.Put(pr)
	assert.Nil(t, err)

	// mock PR not already existing
	d = &memoryStorer{
		prs:    map[string]*api.PublicationReceipt{},
		logger: lg,
		params: params,
	}
	err = d.Put(pr)
	assert.Nil(t, err)
}

func TestMemoryStorer_Put_err(t *testing.T) {
	lg, params := zap.NewNop(), storage.NewDefaultParameters()

	d := &memoryStorer{
		prs:    map[string]*api.PublicationReceipt{},
		logger: lg,
		params: params,
	}
	pr := &api.PublicationReceipt{}
	err := d.Put(pr)
	assert.NotNil(t, err)
}

func TestMemoryStorer_Search_err(t *testing.T) {
	lg, params := zap.NewNop(), storage.NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))

	// invalid search filter
	d := &memoryStorer{
		prs:    map[string]*api.PublicationReceipt{},
		logger: lg,
		params: params,
	}
	fs := &storage.SearchFilters{}
	result, err := d.Search(fs, storage.MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	// too large search limit
	fs = &storage.SearchFilters{
		EntryKey: util.RandBytes(rng, id.Length),
	}
	result, err = d.Search(fs, storage.MaxSearchLimit*2)
	assert.Nil(t, result)
	assert.Equal(t, storage.ErrSearchLimitTooLarge, err)
}
