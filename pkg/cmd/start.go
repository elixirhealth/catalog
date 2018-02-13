package cmd

import (
	"errors"
	"log"
	"os"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/elxirhealth/catalog/pkg/server"
	"github.com/elxirhealth/catalog/pkg/server/storage"
	bserver "github.com/elxirhealth/service-base/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	serverPortFlag       = "serverPort"
	metricsPortFlag      = "metricsPort"
	profilerPortFlag     = "profilerPort"
	profileFlag          = "profile"
	gcpProjectIDFlag     = "gcpProjectID"
	storageInMemoryFlag  = "storageInMemory"
	storageDataStoreFlag = "storageDataStore"
	searchTimeoutFlag    = "searchTimeout"
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
	startCmd.Flags().Bool(storageInMemoryFlag, true,
		"use in-memory storage")
	startCmd.Flags().Bool(storageDataStoreFlag, false,
		"use GCP DataStore storage")
	startCmd.Flags().String(gcpProjectIDFlag, "", "GCP project ID")
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
	storageConfig := &storage.Parameters{
		Type:               storageType,
		SearchQueryTimeout: viper.GetDuration(searchTimeoutFlag),
	}

	c := server.NewDefaultConfig()
	c.WithServerPort(uint(viper.GetInt(serverPortFlag))).
		WithMetricsPort(uint(viper.GetInt(metricsPortFlag))).
		WithProfilerPort(uint(viper.GetInt(profilerPortFlag))).
		WithLogLevel(logging.GetLogLevel(viper.GetString(logLevelFlag))).
		WithProfile(viper.GetBool(profileFlag))
	c.WithStorage(storageConfig).
		WithGCPProjectID(viper.GetString(gcpProjectIDFlag))

	lg := logging.NewDevLogger(c.LogLevel)
	lg.Info("successfully parsed config", zap.Object("config", c))

	return c, nil
}

func getStorageType() (storage.Type, error) {
	if viper.GetBool(storageInMemoryFlag) && viper.GetBool(storageDataStoreFlag) {
		return storage.Unspecified, errMultipleStorageTypes
	}
	if viper.GetBool(storageInMemoryFlag) {
		return storage.InMemory, nil
	}
	if viper.GetBool(storageDataStoreFlag) {
		return storage.DataStore, nil
	}
	return storage.Unspecified, errNoStorageType
}
