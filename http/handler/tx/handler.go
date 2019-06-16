package tx

import (
	"net/http"

	"github.com/lillilli/geth_contract/http/handler"
	"github.com/lillilli/geth_contract/session"

	"github.com/lillilli/geth_contract/eth"
	"github.com/lillilli/logger"
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
	log := logger.NewLogger("tx handler")

	return &Handler{
		log:         log,
		BaseHandler: &handler.BaseHandler{Log: log},

		contractClient:    contractClient,
		userSessionsStore: userSessionsStore,
	}
}

// Get - return tx state
func (h Handler) GetTx(w http.ResponseWriter, r *http.Request) {
	txHash := r.URL.Query().Get("hash")

	tx, exist := h.contractClient.GetTxState(txHash)
	if !exist {
		h.SendBadRequestError(w, "tx not found")
		return
	}

	h.SendMarshalResponse(w, tx)
}

func (h Handler) GetSessionTxs(w http.ResponseWriter, r *http.Request) {
	txs := make([]eth.Tx, 0)
	session := r.URL.Query().Get("session")
	txHashes := h.userSessionsStore.GetWatchedTxs(session)

	for _, txHash := range txHashes {
		tx, exist := h.contractClient.GetTxState(txHash)
		if !exist {
			continue
		}

		txs = append(txs, tx)
	}

	h.SendMarshalResponse(w, txs)
}
