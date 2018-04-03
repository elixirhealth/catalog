// +build acceptance

package acceptance

import (
	"context"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	libriapi "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/client"
	"github.com/elixirhealth/catalog/pkg/server"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/catalog/pkg/server/storage/postgres/migrations"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

type parameters struct {
	nCatalogs      uint
	nAuthorPubKeys uint
	nReaderPubKeys uint
	nPuts          uint
	nSearches      uint
	putTimeout     time.Duration
	searchTimeout  time.Duration
	logLevel       zapcore.Level
}

type state struct {
	catalogs         []*server.Catalog
	catalogClients   []api.CatalogClient
	authorPubKeys    [][]byte
	readerPubKeys    [][]byte
	rng              *rand.Rand
	dbURL            string
	tearDownPostgres func() error
}

func TestAcceptance(t *testing.T) {
	params := &parameters{
		nCatalogs:      3,
		nPuts:          32,
		nSearches:      4,
		nAuthorPubKeys: 4,
		nReaderPubKeys: 4,
		logLevel:       zapcore.InfoLevel,
		putTimeout:     3 * time.Second,
		searchTimeout:  3 * time.Second,
	}
	st := setUp(t, params)

	testPut(t, params, st)

	testSearch(t, params, st)

	tearDown(t, st)
}

func testPut(t *testing.T, params *parameters, st *state) {
	for c := uint(0); c < params.nPuts; c++ {
		i := st.rng.Int31n(int32(params.nAuthorPubKeys))
		j := st.rng.Int31n(int32(params.nReaderPubKeys))
		pr := &api.PublicationReceipt{
			EnvelopeKey:     util.RandBytes(st.rng, id.Length),
			EntryKey:        util.RandBytes(st.rng, id.Length),
			AuthorPublicKey: st.authorPubKeys[i],
			ReaderPublicKey: st.readerPubKeys[j],
			ReceivedTime:    api.ToEpochMicros(time.Now()),
		}
		rq := &api.PutRequest{Value: pr}
		clnt := st.catalogClients[st.rng.Int31n(int32(len(st.catalogClients)))]
		ctx, cancel := context.WithTimeout(context.Background(), params.putTimeout)
		_, err := clnt.Put(ctx, rq)
		cancel()
		assert.Nil(t, err)
	}
}

func testSearch(t *testing.T, params *parameters, st *state) {
	for c := uint(0); c < params.nPuts; c++ {
		i := st.rng.Int31n(int32(params.nReaderPubKeys))
		rq := &api.SearchRequest{
			ReaderPublicKey: st.readerPubKeys[i],
			Limit:           uint32(storage.MaxSearchLimit),
		}
		clnt := st.catalogClients[st.rng.Int31n(int32(len(st.catalogClients)))]
		ctx, cancel := context.WithTimeout(context.Background(), params.searchTimeout)
		rp, err := clnt.Search(ctx, rq)
		cancel()
		assert.Nil(t, err)
		if rp != nil {
			assert.True(t, len(rp.Result) > 0)
			for _, r := range rp.Result {
				assert.Equal(t, rq.ReaderPublicKey, r.ReaderPublicKey)
			}
		}
	}
}

func setUp(t *testing.T, params *parameters) *state {
	rng := rand.New(rand.NewSource(0))
	authorPubKeys := make([][]byte, params.nAuthorPubKeys)
	for i := uint(0); i < params.nAuthorPubKeys; i++ {
		authorPubKeys[i] = util.RandBytes(rng, libriapi.ECPubKeyLength)
	}
	readerPubKeys := make([][]byte, params.nReaderPubKeys)
	for i := uint(0); i < params.nReaderPubKeys; i++ {
		readerPubKeys[i] = util.RandBytes(rng, libriapi.ECPubKeyLength)
	}

	dbURL, cleanup, err := bstorage.StartTestPostgres()
	if err != nil {
		t.Fatal(err)
	}

	st := &state{
		rng:              rng,
		authorPubKeys:    authorPubKeys,
		readerPubKeys:    readerPubKeys,
		dbURL:            dbURL,
		tearDownPostgres: cleanup,
	}
	createAndStartCatalogs(params, st)
	return st
}

func createAndStartCatalogs(params *parameters, st *state) {
	configs, addrs := newCatalogConfigs(params, st)
	catalogs := make([]*server.Catalog, params.nCatalogs)
	catalogClients := make([]api.CatalogClient, params.nCatalogs)
	up := make(chan *server.Catalog, 1)

	for i := uint(0); i < params.nCatalogs; i++ {
		go func() {
			err := server.Start(configs[i], up)
			errors.MaybePanic(err)
		}()

		// wait for server to come up
		catalogs[i] = <-up

		// set up client to it
		var err error
		catalogClients[i], err = client.NewInsecure(addrs[i].String())
		errors.MaybePanic(err)
	}

	st.catalogs = catalogs
	st.catalogClients = catalogClients
}

func newCatalogConfigs(params *parameters, st *state) ([]*server.Config, []*net.TCPAddr) {
	startPort := uint(10100)
	configs := make([]*server.Config, params.nCatalogs)
	addrs := make([]*net.TCPAddr, params.nCatalogs)

	// set eviction params to ensure that evictions actually happen during test
	storageParams := storage.NewDefaultParameters()
	storageParams.Type = bstorage.Postgres
	storageParams.SearchTimeout = params.searchTimeout

	for i := uint(0); i < params.nCatalogs; i++ {
		serverPort, metricsPort := startPort+i*10, startPort+i*10+1
		configs[i] = server.NewDefaultConfig().
			WithStorage(storageParams).
			WithDBUrl(st.dbURL)
		configs[i].WithServerPort(uint(serverPort)).
			WithMetricsPort(uint(metricsPort)).
			WithLogLevel(params.logLevel)
		addrs[i] = &net.TCPAddr{IP: net.ParseIP("localhost"), Port: int(serverPort)}
	}
	return configs, addrs
}

func tearDown(t *testing.T, st *state) {
	logger := &bstorage.ZapLogger{Logger: logging.NewDevInfoLogger()}
	m := bstorage.NewBindataMigrator(
		st.dbURL,
		bindata.Resource(migrations.AssetNames(), migrations.Asset),
		logger,
	)
	err := m.Down()
	assert.Nil(t, err)

	err = st.tearDownPostgres()
	assert.Nil(t, err)
}
