package storage

import (
	"math/rand"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elxirhealth/catalog/pkg/catalogapi"
	"github.com/elxirhealth/service-base/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemoryStorer_PutSearch(t *testing.T) {
	lg := logging.NewDevInfoLogger()
	s := NewMemory(NewDefaultParameters(), lg)
	testStorerPutSearch(t, s)
}

func TestMemoryStorer_Put_ok(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()
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
		storedPRs: map[string]*PublicationReceipt{
			id.Hex(pr.EnvelopeKey): encodeStoredPubReceipt(pr),
		},
		logger: lg,
		params: params,
	}
	err := d.Put(pr)
	assert.Nil(t, err)

	// mock PR not already existing
	d = &memoryStorer{
		storedPRs: map[string]*PublicationReceipt{},
		logger:    lg,
		params:    params,
	}
	err = d.Put(pr)
	assert.Nil(t, err)
}

func TestMemoryStorer_Put_err(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()

	d := &memoryStorer{
		storedPRs: map[string]*PublicationReceipt{},
		logger:    lg,
		params:    params,
	}
	pr := &api.PublicationReceipt{}
	err := d.Put(pr)
	assert.NotNil(t, err)
}

func TestMemoryStorer_Search_err(t *testing.T) {
	lg, params := zap.NewNop(), NewDefaultParameters()
	rng := rand.New(rand.NewSource(0))

	// invalid search filter
	d := &memoryStorer{
		storedPRs: map[string]*PublicationReceipt{},
		logger:    lg,
		params:    params,
	}
	fs := &SearchFilters{}
	result, err := d.Search(fs, MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	// too large search limit
	fs = &SearchFilters{
		EntryKey: util.RandBytes(rng, id.Length),
	}
	result, err = d.Search(fs, MaxSearchLimit*2)
	assert.Nil(t, result)
	assert.Equal(t, ErrSearchLimitTooLarge, err)

	// decodeStoredPubReceipt error
	dummyEnvKey := datastore.NameKey(publicationReceiptKind, "some env key", nil)
	d = &memoryStorer{
		storedPRs: map[string]*PublicationReceipt{
			dummyEnvKey.Name: {
				EnvelopeKey: dummyEnvKey,
			},
		},
		logger: lg,
		params: params,
	}
	result, err = d.Search(fs, MaxSearchLimit)
	assert.Nil(t, result)
	assert.NotNil(t, err)
}
