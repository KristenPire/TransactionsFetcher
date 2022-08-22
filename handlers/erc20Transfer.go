package handlers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"TransactionsFetcher"
)

type TransferHandler struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

type TransferTransaction struct {
	Id           common.Hash
	TxHash       common.Hash
	TokenAddress common.Address
	BlockNumber  big.Int
	BlockHash    common.Hash
	To           common.Address
	From         common.Address
	Kind         string
	Amount       string
	Currency     string
	Raw          string
}

// ToTransaction ...
func (th TransferHandler) ToTransaction(tf TransactionsFetcher.TransactionsFetcher, transaction *types.Transaction) interface{} {
	event, err := tf.Event(transaction.Hash())
	if err != nil {
		return nil
	}
	e, ok := event.(TransferHandler)
	if !ok {
		return nil
	}
	raw, _ := transaction.MarshalJSON()
	kind := "transfer"
	if e.To.Hex() == TransactionsFetcher.EmptyAddress {
		kind = "burn"
	} else if e.From.Hex() == TransactionsFetcher.EmptyAddress {
		kind = "mint"
	}

	return &TransferTransaction{
		Id:           transaction.Hash(),
		TxHash:       transaction.Hash(),
		TokenAddress: tf.Contract(),
		To:           e.To,
		From:         e.From,
		Kind:         kind,
		Amount:       e.Tokens.String(),
		Currency:     "EURE",
		Raw:          string(raw),
	}
}

// ToEvent ...
func (th TransferHandler) ToEvent(tf TransactionsFetcher.TransactionsFetcher, log *types.Log) (interface{}, error) {
	_, topicName := th.Topic()
	var event TransferHandler

	if err := tf.Unpack(&event, topicName, log); err != nil {
		return nil, err
	}
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.To = common.HexToAddress(log.Topics[2].Hex())

	return event, nil
}

// IsRelated ...
func (th TransferHandler) IsRelated(event interface{}, target common.Address) bool {
	e, ok := event.(TransferHandler)
	if !ok {
		return false
	}
	return e.From == target || e.To == target
}

// Topic ...
func (th TransferHandler) Topic() (string, string) {
	return crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex(), "Transfer"
}
