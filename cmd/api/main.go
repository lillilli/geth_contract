package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lillilli/geth_contract/config"
	"github.com/lillilli/geth_contract/eth"
	"github.com/lillilli/geth_contract/http"
	"github.com/lillilli/geth_contract/session"
	"github.com/lillilli/logger"
	"github.com/lillilli/vconf"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

const updateTxsStateInterval = 10 * time.Second

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
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	contractClient, err := eth.NewContractClient(cfg.PrivateKey, cfg.EthNodeURL, cfg.ContractAddress)
	if err != nil {
		return err
	}

	go startUpdateTxsState(ctx, log, contractClient)

	userSessionsStore := session.NewUserSessionStore()
	httpServer := http.NewServer(cfg.HTTP, contractClient, userSessionsStore)

	if err := httpServer.Start(); err != nil {
		return err
	}

	<-signals
	close(signals)

	return httpServer.Stop()
}

func startUpdateTxsState(ctx context.Context, log logger.Logger, contractClient eth.ContractClient) {
	ticker := time.NewTicker(updateTxsStateInterval)

	if err := contractClient.UpdateTxsStates(); err != nil {
		log.Errorf("Updating txs states failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := contractClient.UpdateTxsStates(); err != nil {
				log.Errorf("Updating txs states failed: %v", err)
			}
		}
	}
}
