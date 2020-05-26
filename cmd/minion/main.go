package main

import (
	"detect/log"
	"detect/minion"
	"detect/minion/conf"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var (
	appName = "minion"
)

func newDaemonCommand() *cobra.Command {
	var configPath string
	cmdStart := &cobra.Command{
		Use:   "minion",
		Short: "A distribute task minion",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			agentConfig, err := conf.LoadConfig(configPath, appName)
			if err != nil {
				logrus.Fatalf("parse config error, s%", err)
			}

			logger, err := log.SetLogConf(agentConfig.Log)
			if err != nil {
				logrus.Fatalf("parse config error, s%", err)
			}
			m := minion.NewMinion(logger, agentConfig)
			m.Start()
			listenToSystemSignals(m)
		},
	}
	cmdStart.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
	cmdStart.MarkFlagRequired("config")
	return cmdStart
}

func listenToSystemSignals(m *minion.Minion) {
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
			m.Logger.Debug("System signal: ", sig)
			m.Stop()
			return
		}
	}
}

func main() {
	cmd := newDaemonCommand()
	if err := cmd.Execute(); err != nil {
		logrus.Fatalf("start minion error, s%", err)
		os.Exit(1)
	}
}
