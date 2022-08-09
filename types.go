package TransactionsFetcher

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	EmptyAddress = "0x0000000000000000000000000000000000000000"
)

type TransactionsFetcher interface {
	FetchAll() ([]*Transaction, error)
	Fetch(Query) ([]*Transaction, error)
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
	Tokens *big.Int
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

	logs         []types.Log
	transferLogs map[common.Hash]LogTransfer
	transferTxs  []*Transaction
}
