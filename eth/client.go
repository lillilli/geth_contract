package eth

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	store "github.com/lillilli/geth_contract/contract"
)

// GasLimit - represent tx gas limit
const GasLimit = uint64(300000)

// ContractClient - represents contract client interface
type ContractClient interface {
	GetCount() (big.Int, error)
	Increment() (Tx, error)
	Decrement() (Tx, error)

	GetTxState(hash string) (tx Tx, exist bool)
	UpdateTxsStates() error
}

type contractClient struct {
	*store.Store

	client  *ethclient.Client
	nodeURL string

	publicKeyECDSA *ecdsa.PublicKey
	privateKey     *ecdsa.PrivateKey

	nonce   uint64
	txCache TxCache
	sync.Mutex
}

// NewContractClient - returns new contract client instance
func NewContractClient(privateKeyHash string, nodeURL string, contractAddressHex string) (ContractClient, error) {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, errors.Wrap(err, "open dial connection to node failed")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHash)
	if err != nil {
		return nil, errors.Wrap(err, "parsing private key failed")
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	contractAddress := common.HexToAddress(contractAddressHex)
	accountAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), accountAddress)
	if err != nil {
		return nil, errors.Wrap(err, "getting account nonce failed")
	}

	instance, err := store.NewStore(contractAddress, client)
	if err != nil {
		return nil, errors.Wrap(err, "creating contract client failed")
	}

	return &contractClient{
		Store:          instance,
		client:         client,
		nodeURL:        nodeURL,
		privateKey:     privateKey,
		publicKeyECDSA: publicKeyECDSA,
		nonce:          nonce,
		txCache:        NewTxCache(client),
	}, nil
}

// GetTxState - returns tx state
func (c *contractClient) GetTxState(hash string) (tx Tx, exist bool) {
	return c.txCache.GetTxState(hash)
}

// UpdateTxsStates - updates pending txs states
func (c *contractClient) UpdateTxsStates() error {
	return c.txCache.UpdateStates()
}

// GetCount - returns contract current count field
func (c *contractClient) GetCount() (big.Int, error) {
	count, err := c.Store.GetCount(nil)
	if err != nil {
		return big.Int{}, err
	}

	if count == nil {
		return big.Int{}, errors.New("request returns empty count")
	}

	return *count, nil
}

// Increment - increments contract count field
func (c *contractClient) Increment() (Tx, error) {
	reqData, err := c.counterReqData()
	if err != nil {
		return Tx{}, err
	}

	tx, err := c.Store.IncrementCounter(reqData)
	if err != nil {
		return Tx{}, err
	}

	txData := NewPendingTx(tx.Hash())
	c.txCache.AddPendingTx(txData)

	return txData, err
}

// Increment - decrements contract count field
func (c *contractClient) Decrement() (Tx, error) {
	reqData, err := c.counterReqData()
	if err != nil {
		return Tx{}, err
	}

	tx, err := c.Store.DecrementCounter(reqData)
	if err != nil {
		return Tx{}, err
	}

	txData := NewPendingTx(tx.Hash())
	c.txCache.AddPendingTx(txData)

	return txData, err
}

func (c *contractClient) counterReqData() (*bind.TransactOpts, error) {
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	c.Lock()
	nonce := c.nonce
	c.nonce++
	c.Unlock()

	data := bind.NewKeyedTransactor(c.privateKey)
	data.Nonce = big.NewInt(int64(nonce))
	data.Value = big.NewInt(0) // in wei
	data.GasLimit = GasLimit   // in units
	data.GasPrice = gasPrice

	return data, nil
}
