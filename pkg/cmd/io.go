package cmd

import (
	"context"
	"log"
	"math/rand"
	"time"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/drausin/libri/libri/common/parse"
	api2 "github.com/drausin/libri/libri/librarian/api"
	api "github.com/elixirhealth/catalog/pkg/catalogapi"
	"github.com/elixirhealth/catalog/pkg/client"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	"github.com/elixirhealth/service-base/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	nPutsFlag      = "nPuts"
	nSearchesFlag  = "nSearches"
	timeoutFlag    = "timeout"
	nAuthorPubKeys = 4
	nReaderPubKeys = 4
	logEnvKey      = "envelope_key"
	logNResults    = "n_results"
)

var ioCmd = &cobra.Command{
	Use:   "io",
	Short: "test input/output of one or more catalog servers",
	Run: func(cmd *cobra.Command, args []string) {
		if err := testIO(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	testCmd.AddCommand(ioCmd)

	ioCmd.Flags().Uint(nPutsFlag, 64,
		"number of publications to put into the catalog")
	ioCmd.Flags().Uint(nSearchesFlag, 8,
		"number of searches against the catalog")
	ioCmd.Flags().Uint(timeoutFlag, 3,
		"timeout (secs) of courier requests")

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with prefix
	viper.AutomaticEnv()             // read in environment variables that match
	cerrors.MaybePanic(viper.BindPFlags(ioCmd.Flags()))
}

func testIO() error {
	rng := rand.New(rand.NewSource(0))
	logger := logging.NewDevLogger(logging.GetLogLevel(viper.GetString(logLevelFlag)))
	addrs, err := parse.Addrs(viper.GetStringSlice(catalogsFlag))
	if err != nil {
		return err
	}
	timeout := time.Duration(viper.GetInt(timeoutFlag) * 1e9)
	nPuts := uint(viper.GetInt(nPutsFlag))
	nSearches := uint(viper.GetInt(nSearchesFlag))

	catalogClients := make([]api.CatalogClient, len(addrs))
	for i, addr := range addrs {
		catalogClients[i], err = client.NewInsecure(addr.String())
		if err != nil {
			return err
		}
	}

	authorPubKeys := make([][]byte, nAuthorPubKeys)
	for i := uint(0); i < nAuthorPubKeys; i++ {
		authorPubKeys[i] = util.RandBytes(rng, api2.ECPubKeyLength)
	}
	readerPubKeys := make([][]byte, nReaderPubKeys)
	for i := uint(0); i < nReaderPubKeys; i++ {
		readerPubKeys[i] = util.RandBytes(rng, api2.ECPubKeyLength)
	}

	for c := uint(0); c < nPuts; c++ {
		i := rng.Int31n(int32(nAuthorPubKeys))
		j := rng.Int31n(int32(nReaderPubKeys))
		pr := &api.PublicationReceipt{
			EnvelopeKey:     util.RandBytes(rng, id.Length),
			EntryKey:        util.RandBytes(rng, id.Length),
			AuthorPublicKey: authorPubKeys[i],
			ReaderPublicKey: readerPubKeys[j],
			ReceivedTime:    api.ToEpochMicros(time.Now()),
		}
		rq := &api.PutRequest{Value: pr}
		catClient := catalogClients[rng.Int31n(int32(len(catalogClients)))]
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		_, err := catClient.Put(ctx, rq)
		cancel()
		if err != nil {
			logger.Error("publication put failed", zap.Error(err))
			return err
		}
		logger.Info("publication put succeeded",
			zap.String(logEnvKey, id.Hex(pr.EnvelopeKey)),
		)
	}

	for c := uint(0); c < nSearches; c++ {
		i := rng.Int31n(int32(nReaderPubKeys))
		rq := &api.SearchRequest{
			ReaderPublicKey: readerPubKeys[i],
			Limit:           storage.MaxSearchLimit,
		}
		client := catalogClients[rng.Int31n(int32(len(catalogClients)))]
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		rp, err := client.Search(ctx, rq)
		cancel()
		if err != nil {
			logger.Error("publication search failed", zap.Error(err))
			return err
		}
		logger.Info("found search results", zap.Int(logNResults, len(rp.Result)))
	}
	return nil
}
