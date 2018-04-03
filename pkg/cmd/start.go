package cmd

import (
	"errors"
	"log"
	"os"
	"time"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/elixirhealth/catalog/pkg/server"
	"github.com/elixirhealth/catalog/pkg/server/storage"
	bserver "github.com/elixirhealth/service-base/pkg/server"
	bstorage "github.com/elixirhealth/service-base/pkg/server/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	serverPortFlag      = "serverPort"
	metricsPortFlag     = "metricsPort"
	profilerPortFlag    = "profilerPort"
	profileFlag         = "profile"
	storageMemoryFlag   = "storageMemory"
	searchTimeoutFlag   = "searchTimeout"
	dbURLFlag           = "dbURL"
	storagePostgresFlag = "storagePostgres"
)

var (
	errMultipleStorageTypes = errors.New("multiple storage types specified")
	errNoStorageType        = errors.New("no storage type specified")
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start a catalog server",
	Run: func(cmd *cobra.Command, args []string) {
		writeBanner(os.Stdout)
		time.Sleep(100 * time.Millisecond)
		config, err := getCatalogConfig()
		if err != nil {
			log.Fatal(err)
		}
		if err = server.Start(config, make(chan *server.Catalog, 1)); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().Uint(serverPortFlag, bserver.DefaultServerPort,
		"port for the main service")
	startCmd.Flags().Uint(metricsPortFlag, bserver.DefaultMetricsPort,
		"port for Prometheus metrics")
	startCmd.Flags().Uint(profilerPortFlag, bserver.DefaultProfilerPort,
		"port for profiler endpoints (when enabled)")
	startCmd.Flags().Bool(profileFlag, bserver.DefaultProfile,
		"whether to enable profiler")
	startCmd.Flags().Bool(storageMemoryFlag, false,
		"use in-memory storage")
	startCmd.Flags().Bool(storagePostgresFlag, false,
		"use Postgres DB storage")
	startCmd.Flags().String(dbURLFlag, "", "Postgres DB URL")
	startCmd.Flags().Duration(searchTimeoutFlag, storage.DefaultSearchQueryTimeout,
		"timeout for Search DataStore requests")

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with "COURIER_" prefix
	viper.AutomaticEnv()             // read in environment variables that match
	cerrors.MaybePanic(viper.BindPFlags(startCmd.Flags()))
}

func getCatalogConfig() (*server.Config, error) {
	storageType, err := getStorageType()
	if err != nil {
		return nil, err
	}

	storageConfig := storage.NewDefaultParameters()
	storageConfig.Type = storageType
	storageConfig.SearchTimeout = viper.GetDuration(searchTimeoutFlag)

	c := server.NewDefaultConfig()
	c.WithServerPort(uint(viper.GetInt(serverPortFlag))).
		WithMetricsPort(uint(viper.GetInt(metricsPortFlag))).
		WithProfilerPort(uint(viper.GetInt(profilerPortFlag))).
		WithLogLevel(logging.GetLogLevel(viper.GetString(logLevelFlag))).
		WithProfile(viper.GetBool(profileFlag))
	c.WithStorage(storageConfig).
		WithDBUrl(viper.GetString(dbURLFlag))

	lg := logging.NewDevLogger(c.LogLevel)
	lg.Info("successfully parsed config", zap.Object("config", c))

	return c, nil
}

func getStorageType() (bstorage.Type, error) {
	if viper.GetBool(storageMemoryFlag) && viper.GetBool(storagePostgresFlag) {
		return bstorage.Unspecified, errMultipleStorageTypes
	}
	if viper.GetBool(storageMemoryFlag) {
		return bstorage.Memory, nil
	}
	if viper.GetBool(storagePostgresFlag) {
		return bstorage.Postgres, nil
	}
	if viper.GetBool(storagePostgresFlag) {
		return bstorage.DataStore, nil
	}
	return bstorage.Unspecified, errNoStorageType
}
