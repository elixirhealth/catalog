package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	errors2 "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/catalog/pkg/server/storage/postgres/migrations"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func TestStorer_Put_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	lg := logging.NewDevLogger(zapcore.DebugLevel)
	params := storage.NewDefaultParameters()
	pr := okPubReceipt(rng)

	var err error
	var s *storer

	s = &storer{
		params: params,
		logger: lg,
		qr: &fixedQuerier{
			insertResult: &fixedSqlResult{
				rowsAffectedValue: 1, // new PR
			},
		},
	}
	err = s.Put(pr)
	assert.Nil(t, err)

	s = &storer{
		params: params,
		logger: lg,
		qr: &fixedQuerier{
			insertResult: &fixedSqlResult{
				rowsAffectedValue: 0, // existing PR
			},
		},
	}
	err = s.Put(pr)
	assert.Nil(t, err)
}

func TestStorer_Put_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	lg := zap.NewNop() // logging.NewDevLogger(zapcore.DebugLevel)
	params := storage.NewDefaultParameters()
	cases := map[string]struct {
		s        *storer
		pr       *api.PublicationReceipt
		expected error
	}{
		"bad PR": {
			s: &storer{
				params: params,
				logger: lg,
			},
			pr:       &api.PublicationReceipt{},
			expected: api.ErrMissingEnvelopeKey,
		},
		"insert err": {
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					insertErr: errTest,
				},
			},
			pr:       okPubReceipt(rng),
			expected: errTest,
		},
	}
	for desc, c := range cases {
		err := c.s.Put(c.pr)
		assert.Equal(t, err, c.expected, desc)
	}
}

func TestStorer_Search_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	lg := zap.NewNop() // logging.NewDevLogger(zapcore.DebugLevel)
	params := storage.NewDefaultParameters()

	prs1 := []*api.PublicationReceipt{
		okPubReceipt(rng),
		okPubReceipt(rng),
	}
	var s *storer
	var fs *storage.SearchFilters
	var prs2 []*api.PublicationReceipt
	var err error

	// some results
	s = &storer{
		params: params,
		logger: lg,
		qr: &fixedQuerier{
			selectRows: &fixedQueryRows{
				prs: prs1,
			},
		},
	}
	fs = &storage.SearchFilters{EntryKey: prs1[0].EntryKey}
	prs2, err = s.Search(fs, storage.MaxSearchLimit)
	assert.Nil(t, err)
	assert.Equal(t, prs1, prs2)

	// no results
	s = &storer{
		params: params,
		logger: lg,
		qr: &fixedQuerier{
			selectErr: sql.ErrNoRows,
		},
	}
	prs2, err = s.Search(fs, storage.MaxSearchLimit)
	assert.Nil(t, err)
	assert.Len(t, prs2, 0)
}

func TestStorer_Search_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	lg := zap.NewNop() // logging.NewDevLogger(zapcore.DebugLevel)
	params := storage.NewDefaultParameters()
	prs := []*api.PublicationReceipt{
		okPubReceipt(rng),
		okPubReceipt(rng),
	}
	okFilters := &storage.SearchFilters{
		EntryKey: prs[0].EntryKey,
	}

	cases := map[string]struct {
		s        *storer
		fs       *storage.SearchFilters
		limit    uint32
		expected error
	}{
		"bad filters": {
			s: &storer{
				params: params,
				logger: lg,
			},
			fs:       &storage.SearchFilters{},
			limit:    128,
			expected: storage.ErrNoEqualityFilters,
		},
		"bad limit": {
			s: &storer{
				params: params,
				logger: lg,
			},
			fs:       okFilters,
			limit:    129,
			expected: storage.ErrSearchLimitTooLarge,
		},
		"select query err": {
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectErr: errTest,
				},
			},
			fs:       okFilters,
			limit:    128,
			expected: errTest,
		},
		"scan err": {
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectRows: &fixedQueryRows{
						prs:     prs,
						scanErr: errTest,
					},
				},
			},
			fs:       okFilters,
			limit:    128,
			expected: errTest,
		},
		"results close err": {
			s: &storer{
				params: params,
				logger: lg,
				qr: &fixedQuerier{
					selectRows: &fixedQueryRows{
						prs:      prs,
						closeErr: errTest,
					},
				},
			},
			fs:       okFilters,
			limit:    128,
			expected: errTest,
		},
	}
	for desc, c := range cases {
		prs, err := c.s.Search(c.fs, c.limit)
		assert.Equal(t, c.expected, err, desc)
		assert.Nil(t, prs, desc)
	}
}

func okPubReceipt(rng *rand.Rand) *api.PublicationReceipt {
	return &api.PublicationReceipt{
		EnvelopeKey:     util.RandBytes(rng, id.Length),
		EntryKey:        util.RandBytes(rng, id.Length),
		AuthorPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReaderPublicKey: util.RandBytes(rng, libriapi.ECPubKeyLength),
		ReceivedTime:    api.ToEpochMicros(time.Now()),
	}
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

func TestGetSearchRangePreds(t *testing.T) {
	now := time.Now()
	nowMicros := now.UnixNano() / 1E3
	tomorrow := now.Add(24 * time.Hour).Format(isoDate)
	earlier := time.Now().Add(-1 * time.Hour)
	earlierMicros := earlier.UnixNano() / 1E3
	yesterday := earlier.Add(-24 * time.Hour).Format(isoDate)

	var fs *storage.SearchFilters
	var preds string

	fs = &storage.SearchFilters{}
	preds = getSearchRangePreds(fs)
	assert.Len(t, preds, 0)

	fs = &storage.SearchFilters{
		Before: nowMicros,
	}
	preds = getSearchRangePreds(fs)
	expected := fmt.Sprintf("received_time < %d AND received_date <= '%s'", nowMicros, tomorrow)
	assert.Equal(t, expected, preds)

	fs = &storage.SearchFilters{
		Before: nowMicros,
		After:  earlierMicros,
	}
	preds = getSearchRangePreds(fs)
	expected = strings.Join([]string{
		fmt.Sprintf("received_time < %d", nowMicros),
		fmt.Sprintf("received_date <= '%s'", tomorrow),
		fmt.Sprintf("received_time >= %d", earlierMicros),
		fmt.Sprintf("received_date >= '%s'", yesterday),
	}, " AND ")
	assert.Equal(t, expected, preds)

}

type fixedQuerier struct {
	selectRows   QueryRows
	selectErr    error
	insertResult sql.Result
	insertErr    error
}

func (f *fixedQuerier) SelectQueryContext(ctx context.Context, b squirrel.SelectBuilder) (QueryRows, error) {
	return f.selectRows, f.selectErr
}

func (f *fixedQuerier) SelectQueryRowContext(ctx context.Context, b squirrel.SelectBuilder) squirrel.RowScanner {
	panic("implement me")
}

func (f *fixedQuerier) InsertExecContext(ctx context.Context, b squirrel.InsertBuilder) (sql.Result, error) {
	return f.insertResult, f.insertErr
}

func (f *fixedQuerier) UpdateExecContext(ctx context.Context, b squirrel.UpdateBuilder) (sql.Result, error) {
	panic("implement me")
}

type fixedSqlResult struct {
	rowsAffectedValue int64
	rowsAffectedErr   error
}

func (f *fixedSqlResult) LastInsertId() (int64, error) {
	panic("implement me")
}

func (f *fixedSqlResult) RowsAffected() (int64, error) {
	return f.rowsAffectedValue, f.rowsAffectedErr
}

type fixedQueryRows struct {
	prs      []*api.PublicationReceipt
	scanErr  error
	errErr   error
	cursor   int
	closeErr error
}

func (f *fixedQueryRows) Scan(dest ...interface{}) error {
	if f.scanErr != nil {
		return f.scanErr
	}
	pr := f.prs[f.cursor]
	dest[0] = &pr.EnvelopeKey
	dest[1] = &pr.EntryKey
	dest[2] = &pr.AuthorPublicKey
	dest[3] = &pr.AuthorEntityId
	dest[4] = &pr.ReaderPublicKey
	dest[5] = &pr.ReaderEntityId
	dest[6] = &pr.ReceivedTime
	f.cursor++
	return nil
}

func (f *fixedQueryRows) Next() bool {
	return f.cursor < len(f.prs)
}

func (f *fixedQueryRows) Close() error {
	return f.closeErr
}

func (f *fixedQueryRows) Err() error {
	return f.errErr
}
