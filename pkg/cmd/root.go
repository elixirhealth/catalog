package cmd

import (
	"log"

	"github.com/drausin/libri/libri/common/errors"
	bserver "github.com/elxirhealth/service-base/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envVarPrefix = "CATALOG"
	logLevelFlag = "logLevel"
)

var rootCmd = &cobra.Command{
	Short: "TODO", // TODO
}

func init() {
	rootCmd.PersistentFlags().String(logLevelFlag, bserver.DefaultLogLevel.String(),
		"log level")

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with given prefix
	viper.AutomaticEnv()             // read in environment variables that match
	errors.MaybePanic(viper.BindPFlags(rootCmd.Flags()))
}

// Execute runs the root catalog command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
