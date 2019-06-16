package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/lillilli/geth_contract/session"

	"github.com/gorilla/mux"
	"github.com/lillilli/logger"
	"github.com/pkg/errors"

	"github.com/lillilli/geth_contract/config"
	"github.com/lillilli/geth_contract/eth"

	stateHandler "github.com/lillilli/geth_contract/http/handler/state"
	txHandler "github.com/lillilli/geth_contract/http/handler/tx"
)

const (
	readTimeout  = 20 * time.Second
	writeTimeout = readTimeout
)

// Server - http server interface
type Server interface {
	Start() error
	Stop() error

	Address() string
}

// server - http sever structure
type server struct {
	contractClient    eth.ContractClient
	userSessionsStore session.UserSessionStore

	server *http.Server
	mux    *mux.Router

	log logger.Logger
}

// NewServer - return new instance of http server
func NewServer(cfg config.HTTPServer, contractClient eth.ContractClient, userSessionsStore session.UserSessionStore) Server {
	api := &server{
		contractClient:    contractClient,
		userSessionsStore: userSessionsStore,
		mux:               mux.NewRouter(),
		log:               logger.NewLogger("http server"),
	}

	api.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      api.mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return api
}

// Start - start http server
func (s server) Start() error {
	s.log.Info("Starting...")

	s.declareRoutes()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", s.server.Addr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to get address %s", s.server.Addr))
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to start listening %s", tcpAddr))
	}

	go func() {
		err := s.server.Serve(listener)
		s.log.Errorf("Serving error: %v", err)
	}()

	s.log.Infof("Listening on %s", s.server.Addr)
	return nil
}

func (s server) declareRoutes() {
	stateHandler := stateHandler.New(s.contractClient, s.userSessionsStore)
	txHandler := txHandler.New(s.contractClient, s.userSessionsStore)

	s.mux.HandleFunc("/state/latest", enableCORS(stateHandler.Latest)).Methods(http.MethodGet)
	s.mux.HandleFunc("/state/increment", enableCORS(stateHandler.Increment)).Methods(http.MethodGet)
	s.mux.HandleFunc("/state/decrement", enableCORS(stateHandler.Decrement)).Methods(http.MethodGet)

	s.mux.HandleFunc("/tx/state", enableCORS(txHandler.GetTx)).Methods(http.MethodGet)
	s.mux.HandleFunc("/tx/session", enableCORS(txHandler.GetSessionTxs)).Methods(http.MethodGet)
}

// Stop - shutdown http server
func (s server) Stop() error {
	s.log.Info("Stopping...")
	return s.server.Shutdown(context.TODO())
}

// Address - return server address
func (s server) Address() string {
	return s.server.Addr
}
