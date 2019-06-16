package eth

import (
	"github.com/ethereum/go-ethereum/common"
)

type Tx struct {
	Hash    common.Hash `json:"hash"`
	Pending bool        `json:"pending"`
}

func NewPendingTx(hash common.Hash) Tx {
	return Tx{
		Hash:    hash,
		Pending: true,
	}
}
