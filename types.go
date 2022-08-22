package TransactionsFetcher

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	EmptyAddress  = "0x0000000000000000000000000000000000000000"
	BlockRangeMax = 3500
)

type TransactionsFetcher interface {
	Fetch(Query) ([]interface{}, error)
	Unpack(event interface{}, topicName string, log *types.Log) error
	Event(Id common.Hash) (interface{}, error)
	Contract() common.Address
}

type TransactionsFetcherHandler interface {
	ToTransaction(TransactionsFetcher, *types.Transaction) interface{}
	ToEvent(TransactionsFetcher, *types.Log) (interface{}, error)
	IsRelated(event interface{}, target common.Address) bool
	Topic() (string, string)
}

type Query struct {
	FromBlock *big.Int
	ToBlock   *big.Int
	Target    common.Address
	Limit     uint64
}

type LogTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
}

type Transaction struct {
	Id           common.Hash
	TxHash       common.Hash
	Address      common.Address // ?
	TokenAddress common.Address
	BlockNumber  big.Int
	BlockHash    common.Hash
	ToAddress    common.Address
	FromAddress  common.Address
	Kind         string
	Amount       string
	Currency     string
	Raw          string
}

type transactionsFetcher struct {
	client      *ethclient.Client
	contract    common.Address
	target      common.Address
	contractABI abi.ABI
	handler     TransactionsFetcherHandler

	logs         []types.Log
	eventLogs    map[common.Hash]interface{}
	transactions []interface{}
}
