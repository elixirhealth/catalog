package postgres

import (
	"errors"
	"math/rand"
	"testing"

	errors2 "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/catalog/pkg/server/storage/postgres/migrations"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

var (
	errTest = errors.New("test error")
)

func setUpPostgresTest() (string, func() error) {
	dbURL, cleanup, err := bstorage.StartTestPostgres()
	errors2.MaybePanic(err)
	as := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	logger := &bstorage.LogLogger{}
	m := bstorage.NewBindataMigrator(dbURL, as, logger)
	errors2.MaybePanic(m.Up())
	return dbURL, func() error {
		if err := m.Down(); err != nil {
			return err
		}
		return cleanup()
	}
}

// TestStorer_PutSearch is an integration-style test that spins up a whole DB, puts a bunch of
// publications, and then searches for them.
func TestStorer_PutSearch(t *testing.T) {
	dbURL, tearDown := setUpPostgresTest()
	defer func() {
		err := tearDown()
		assert.Nil(t, err)
	}()

	lg := logging.NewDevLogger(zapcore.DebugLevel)
	params := storage.NewDefaultParameters()
	params.Type = bstorage.Postgres
	s, err := New(dbURL, params, lg)
	assert.Nil(t, err)
	assert.NotNil(t, s)

	storage.TestStorerPutSearch(t, s)
}

func TestGetSearchEqPreds(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	var fs *storage.SearchFilters
	var preds map[string]interface{}

	fs = &storage.SearchFilters{}
	preds = getSearchEqPreds(fs)
	assert.Len(t, preds, 0)

	fs = &storage.SearchFilters{
		EntryKey:        util.RandBytes(rng, 32),
		AuthorPublicKey: util.RandBytes(rng, 33),
		AuthorEntityID:  "author entity ID",
		ReaderPublicKey: util.RandBytes(rng, 33),
		ReaderEntityID:  "reader entity ID",
	}
	preds = getSearchEqPreds(fs)
	assert.Equal(t, fs.EntryKey, preds[entryKeyCol])
	assert.Equal(t, fs.AuthorPublicKey, preds[authorPubKeyCol])
	assert.Equal(t, fs.AuthorEntityID, preds[authorEntityIDCol])
	assert.Equal(t, fs.ReaderPublicKey, preds[readerPubKeyCol])
	assert.Equal(t, fs.ReaderEntityID, preds[readerEntityIDCol])
}
