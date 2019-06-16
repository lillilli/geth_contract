package eth

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lillilli/logger"
)

// TxCache - represents tx cache interface
type TxCache interface {
	AddPendingTx(tx Tx)
	GetTxState(hash string) (tx Tx, exist bool)

	UpdateStates() error
}

type txCache struct {
	client *ethclient.Client

	latestBlock uint64
	finishedTxs map[string]Tx
	pendingTxs  map[string]Tx

	log logger.Logger
	sync.RWMutex
}

// NewTxCache - returns new tx cache instance
func NewTxCache(client *ethclient.Client) TxCache {
	return &txCache{
		client:      client,
		pendingTxs:  make(map[string]Tx),
		finishedTxs: make(map[string]Tx),
		log:         logger.NewLogger("tx cache"),
	}
}

// AddPendingTx - adds pending tx to watched list
func (c *txCache) AddPendingTx(tx Tx) {
	c.Lock()
	c.pendingTxs[tx.Hash.String()] = tx
	c.Unlock()
}

// GetTxState - returns tx state
func (c *txCache) GetTxState(hash string) (tx Tx, exist bool) {
	c.RLock()
	defer c.RUnlock()

	tx, ok := c.finishedTxs[hash]
	if ok {
		return tx, ok
	}

	tx, ok = c.pendingTxs[hash]
	return tx, ok
}

// UpdateStates - updates pending txs states
func (c *txCache) UpdateStates() error {
	c.log.Debug("Updating txs statuses....")

	latestBlock, err := c.client.BlockByNumber(context.TODO(), nil)
	if err != nil {
		return err
	}

	if c.getLatestBlock() == latestBlock.NumberU64() {
		return nil
	}

	c.updateLatestBlock(latestBlock.NumberU64())

	pendingTxsCopy := c.pendingTxsCopy()
	finishedTxsCopy := c.finishedTxsCopy()

	for hash, tx := range pendingTxsCopy {
		_, pending, err := c.client.TransactionByHash(context.TODO(), tx.Hash)
		if err != nil {
			return err
		}

		if tx.Pending = pending; !tx.Pending {
			c.log.Debugf("Find finished tx with hash %q", hash)
			delete(pendingTxsCopy, hash)
			finishedTxsCopy[hash] = tx
		}
	}

	c.Lock()
	c.pendingTxs = pendingTxsCopy
	c.finishedTxs = finishedTxsCopy
	c.Unlock()

	return nil
}

func (c *txCache) getLatestBlock() uint64 {
	c.RLock()
	defer c.RUnlock()

	return c.latestBlock
}

func (c *txCache) updateLatestBlock(v uint64) {
	c.Lock()
	c.latestBlock = v
	c.Unlock()
}

func (c *txCache) pendingTxsCopy() map[string]Tx {
	mapCopy := make(map[string]Tx)

	c.RLock()
	for k, v := range c.pendingTxs {
		mapCopy[k] = v
	}
	c.RUnlock()

	return mapCopy
}

func (c *txCache) finishedTxsCopy() map[string]Tx {
	mapCopy := make(map[string]Tx)

	c.RLock()
	for k, v := range c.finishedTxs {
		mapCopy[k] = v
	}
	c.RUnlock()

	return mapCopy
}
