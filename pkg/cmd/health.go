package cmd

import (
	"log"
	"net"
	"os"

	lserver "github.com/drausin/libri/libri/common/logging"
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "test health of one or more catalog servers",
	Run: func(cmd *cobra.Command, args []string) {
		hc, err := getHealthChecker()
		if err != nil {
			log.Fatal(err)
		}
		if allOk, _ := hc.Check(); !allOk {
			os.Exit(1)
		}
	},
}

func init() {
	testCmd.AddCommand(healthCmd)
}

func getHealthChecker() (server.HealthChecker, error) {
	addrs, err := parseAddrs(viper.GetStringSlice(catalogsFlag))
	if err != nil {
		return nil, err
	}
	lg := lserver.NewDevLogger(lserver.GetLogLevel(viper.GetString(logLevelFlag)))
	return server.NewHealthChecker(server.NewInsecureDialer(), addrs, lg)
}

// parseAddrs parses an array of net.TCPAddrs from an array of IPv4:Port address strings.
// TODO (drausin) move to libri lib once separated from server pkg
func parseAddrs(addrs []string) ([]*net.TCPAddr, error) {
	netAddrs := make([]*net.TCPAddr, len(addrs))
	for i, a := range addrs {
		netAddr, err := net.ResolveTCPAddr("tcp4", a)
		if err != nil {
			return nil, err
		}
		netAddrs[i] = netAddr
	}
	return netAddrs, nil
}
