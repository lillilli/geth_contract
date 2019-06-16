package handler

import (
	"encoding/json"
	"net/http"

	"github.com/lillilli/logger"
)

const (
	internalErrPrefix = "internal error: "
)

// BaseHandler - base http handler
type BaseHandler struct {
	Log logger.Logger
}

// SendBadRequestError - send http response with 400 status and specified message
func (h BaseHandler) SendBadRequestError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)

	if _, err := w.Write([]byte(message)); err != nil {
		h.Log.Errorf("Error while sending bad request: %v", err)
	}
}

// SendInternalError - send http response with 500 status and specified message
func (h BaseHandler) SendInternalError(w http.ResponseWriter, err error, msg string) {
	h.Log.Errorf("Handle error (%s): %v", msg, err)

	w.WriteHeader(http.StatusInternalServerError)

	if _, err = w.Write([]byte(internalErrPrefix + msg)); err != nil {
		h.Log.Errorf("Error while sending internal error: %v", err)
	}
}

// SendMarshalResponse - send http response with 200 status and stringified specified data
func (h BaseHandler) SendMarshalResponse(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		h.SendInternalError(w, err, "parsing output data failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(b); err != nil {
		h.Log.Errorf("Error while sending response: %v", err)
	}
}
