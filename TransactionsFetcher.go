package TransactionsFetcher

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// New ...
func New(client *ethclient.Client, contract common.Address, contractABI string) (TransactionsFetcher, error) {
	ca, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}
	return &transactionsFetcher{
		client:       client,
		contract:     contract,
		contractABI:  ca,
		transferLogs: make(map[common.Hash]LogTransfer),
	}, nil
}

// FetchAll ...
func (tf *transactionsFetcher) FetchAll() ([]*Transaction, error) {
	lastBlockNumber, err := tf.getLastBlockNumber()
	if err != nil {
		return nil, err
	}
	if err := tf.newQuery(big.NewInt(0), lastBlockNumber); err != nil {
		fmt.Println("Query issue")
		return nil, err
	}
	if err := tf.fetchTranferLogs(common.Address{}, false); err != nil {
		fmt.Println("error in fetchTransferLogs")
		return nil, err
	}
	if err := tf.fetchTransferTransactions(0); err != nil {
		fmt.Println("error in fetchTransferTransactions")
		return nil, err
	}
	return tf.transferTxs, nil
}

// Fetch ...
func (tf *transactionsFetcher) Fetch(q Query) ([]*Transaction, error) {
	if err := tf.validateQuery(&q); err != nil {
		return nil, err
	}
	if err := tf.newQuery(q.FromBlock, q.ToBlock); err != nil {
		fmt.Println("Query issue")
		return nil, err
	}
	if err := tf.fetchTranferLogs(q.Target, true); err != nil {
		fmt.Println("error in fetchTransferLogs")
		return nil, err
	}
	if err := tf.fetchTransferTransactions(q.Limit); err != nil {
		fmt.Println("error in fetchTransferTransactions")
		return nil, err
	}
	return tf.transferTxs, nil
}

// validateQuery ...
func (tf transactionsFetcher) validateQuery(q *Query) error {
	if q.FromBlock == nil {
		q.FromBlock = big.NewInt(0)
	}
	if q.ToBlock == nil {
		var err error
		q.ToBlock, err = tf.getLastBlockNumber()
		if err != nil {
			return err
		}
	}
	return nil
}

func (tf *transactionsFetcher) newQuery(from *big.Int, to *big.Int) error {
	query := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{
			tf.contract,
		},
	}
	var err error
	tf.logs, err = tf.client.FilterLogs(context.Background(), query)
	return err
}

func (tf *transactionsFetcher) fetchTranferLogs(target common.Address, withTarget bool) error {
	topicSignature, topicName := ERC20.Transfer()
	for _, vLog := range tf.logs {
		switch vLog.Topics[0].Hex() {
		case topicSignature:
			//Topic specific code
			var transferEvent LogTransfer
			if err := tf.contractABI.UnpackIntoInterface(&transferEvent, topicName, vLog.Data); err != nil {
				return err
			}
			transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
			//End of Topic specific code
			if withTarget && tf.isRelated(transferEvent, target) == false {
				continue
			}
			tf.transferLogs[vLog.TxHash] = transferEvent
		}
	}
	return nil
}

func (tf *transactionsFetcher) fetchTransferTransactions(limit uint64) error {
	fetched := uint64(0)
	for txID, _ := range tf.transferLogs {
		tx, _, err := tf.client.TransactionByHash(context.Background(), txID)
		if err != nil {
			return err
		}
		tf.transferTxs = append(tf.transferTxs, tf.toTransaction(tx))
		fetched++
		if limit > 0 && fetched >= limit {
			return nil
		}
	}
	return nil
}

func (tf *transactionsFetcher) toTransaction(tx *types.Transaction) *Transaction {
	raw, _ := tx.MarshalJSON() // Don't let it like this
	kind  := "transfer"
	if tf.transferLogs[tx.Hash()].To.Hex() == EmptyAddress {
		kind = "burn"
	} else if tf.transferLogs[tx.Hash()].From.Hex() == EmptyAddress {
		kind = "mint"
	}

	return &Transaction{
		Id:           tx.Hash(),
		TxHash:       tx.Hash(),
		TokenAddress: tf.contract,
		ToAddress:    tf.transferLogs[tx.Hash()].To,
		FromAddress:  tf.transferLogs[tx.Hash()].From,
		Kind:         kind,
		Amount:       tf.transferLogs[tx.Hash()].Tokens.String(),
		Currency: "EURE",
		Raw:      string(raw),
	}
}

func (tf transactionsFetcher) isRelated(te LogTransfer, t common.Address) bool {
	return te.From == t || te.To == t
}

func (tf transactionsFetcher) getLastBlockNumber() (*big.Int, error) {
	header, err := tf.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
}
