package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lillilli/geth_contract/config"
	"github.com/lillilli/geth_contract/eth"
	"github.com/lillilli/geth_contract/http"
	"github.com/lillilli/geth_contract/session"
	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
	"github.com/robfig/cron"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Log)
	log := logger.NewLogger("api")

	if err := runService(cfg, log); err != nil {
		log.Errorf("Run service error: %v", err)
		os.Exit(1)
	}
}

func runService(cfg *config.Config, log logger.Logger) error {
	cronInstance := cron.New()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	cronInstance.Start()
	defer cronInstance.Stop()

	contractClient, err := eth.NewContractClient(cfg.PrivateKey, cfg.EthNodeURL, cfg.ContractAddress)
	if err != nil {
		return err
	}

	if err := startUpdateTxsState(cronInstance, log, contractClient); err != nil {
		return err
	}

	userSessionsStore := session.NewUserSessionStore()
	httpServer := http.NewServer(cfg.HTTP, contractClient, userSessionsStore)

	if err := httpServer.Start(); err != nil {
		return err
	}

	<-signals
	close(signals)

	return httpServer.Stop()
}

func startUpdateTxsState(cronInstance *cron.Cron, log logger.Logger, contractClient eth.ContractClient) error {
	if err := contractClient.UpdateTxsStates(); err != nil {
		log.Errorf("Updating txs states failed: %v", err)
	}

	return cronInstance.AddFunc("*/10 * * * * *", func() {
		if err := contractClient.UpdateTxsStates(); err != nil {
			log.Errorf("Updating txs states failed: %v", err)
		}
	})
}
