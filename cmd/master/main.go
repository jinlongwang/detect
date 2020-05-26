package main

import (
	"detect/log"
	"detect/master"
	"detect/master/conf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var (
	appName = "master"
)

func newDaemonCommand() *cobra.Command {
	var configPath string
	cmdStart := &cobra.Command{
		Use:   "master",
		Short: "A distribute task master",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			masterConfig, err := conf.LoadConfig(configPath, appName)
			if err != nil {
				logrus.Fatalf("parse config error, s%", err)
			}

			logger, err := log.SetLogConf(masterConfig.Log)
			if err != nil {
				logrus.Fatalf("parse config error, s%", err)
			}
			m := master.NewMaster(logger, masterConfig)
			m.Start()
			listenToSystemSignals(m)
		},
	}
	cmdStart.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
	cmdStart.MarkFlagRequired("config")
	return cmdStart
}

func listenToSystemSignals(m *master.Master) {
	signalChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sighupChan:
			m.Logger.Debug("receive hup signal")
			m.Stop()
		case sig := <-signalChan:
			m.Logger.Debug("System signal: %s", sig)
			m.Stop()
			return
		}
	}
}

func main() {
	cmd := newDaemonCommand()
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("start master error, s%", err)
		os.Exit(1)
	}
}
