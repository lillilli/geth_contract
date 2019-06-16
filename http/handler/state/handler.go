package state

import (
	"net/http"

	"github.com/lillilli/logger"

	"github.com/lillilli/geth_contract/eth"
	"github.com/lillilli/geth_contract/http/handler"
	"github.com/lillilli/geth_contract/session"
)

// Handler - http handler structure
type Handler struct {
	*handler.BaseHandler

	contractClient    eth.ContractClient
	userSessionsStore session.UserSessionStore

	log logger.Logger
}

// New - return new handler instance
func New(contractClient eth.ContractClient, userSessionsStore session.UserSessionStore) *Handler {
	log := logger.NewLogger("state handler")

	return &Handler{
		log:         log,
		BaseHandler: &handler.BaseHandler{Log: log},

		contractClient:    contractClient,
		userSessionsStore: userSessionsStore,
	}
}

// Latest - return current system state
func (h Handler) Latest(w http.ResponseWriter, r *http.Request) {
	count, err := h.contractClient.GetCount()
	if err != nil {
		h.SendInternalError(w, err, "getting current state failed")
		return
	}

	resp := SystemStateResp{Value: count.String()}
	h.SendMarshalResponse(w, resp)
}

// Increment - increment current system state
func (h Handler) Increment(w http.ResponseWriter, r *http.Request) {
	session := r.URL.Query().Get("session")
	h.log.Debugf("Incrementing system state by req (session = %q)", session)

	tx, err := h.contractClient.Increment()
	if err != nil {
		h.SendInternalError(w, err, "getting current state failed")
		return
	}

	h.userSessionsStore.AddWatchedTx(session, tx.Hash.String())
	h.SendMarshalResponse(w, tx)
}

// Decrement - decrement current system state
func (h Handler) Decrement(w http.ResponseWriter, r *http.Request) {
	session := r.URL.Query().Get("session")
	h.log.Debugf("Decrementing system state by req (session = %q)", session)

	tx, err := h.contractClient.Decrement()
	if err != nil {
		h.SendInternalError(w, err, "getting current state failed")
		return
	}

	h.userSessionsStore.AddWatchedTx(session, tx.Hash.String())
	h.SendMarshalResponse(w, tx)
}
