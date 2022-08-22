package TransactionsFetcher

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// New ...
func New(client *ethclient.Client, contract common.Address, contractABI string, handler TransactionsFetcherHandler) (TransactionsFetcher, error) {
	ca, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}
	return &transactionsFetcher{
		client:       client,
		contract:     contract,
		contractABI:  ca,
		handler:      handler,
		eventLogs:    make(map[common.Hash]interface{}),
	}, nil
}

// PUBLIC
// Fetch ...
func (tf *transactionsFetcher) Fetch(q Query) ([]interface{}, error) {
	t := time.Now()
	if err := tf.validateQuery(&q); err != nil {
		return nil, err
	}
	if err := tf.newQuery(q.FromBlock, q.ToBlock); err != nil {
		fmt.Println("Query issue")
		return nil, err
	}
	fmt.Println("time for newQuery %s", time.Since(t))
	if err := tf.fetchLogs(q.Target, true); err != nil {
		fmt.Println("error in fetchLogs")
		return nil, err
	}
	fmt.Println("time for fetchTransferLogs %s", time.Since(t))
	if err := tf.fetchTransactions(q.Limit); err != nil {
		fmt.Println("error in fetchTransferTransactions")
		return nil, err
	}
	fmt.Println("time for fetchTranferTransactions %s", time.Since(t))
	return tf.transactions, nil
}

// Event ...
func (tf *transactionsFetcher) Event(id common.Hash) (interface{}, error) {
	event, ok := tf.eventLogs[id]
	if !ok {
		return nil, fmt.Errorf("Error: id (%s) not found", id.Hex())
	}
	return event, nil
}

// Contract
func (tf transactionsFetcher) Contract() common.Address {
	return tf.contract
}

// Unpack
func (tf *transactionsFetcher) Unpack(event interface{}, topicName string, log *types.Log) error {
	return tf.contractABI.UnpackIntoInterface(event, topicName, log.Data)
}

// PRIVATE
// getLastBlockNumber ...
func (tf transactionsFetcher) getLastBlockNumber() (*big.Int, error) {
	header, err := tf.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return header.Number, nil
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
// newQuerySafe ...
func (tf *transactionsFetcher) newQuerySafe(from *big.Int, to *big.Int) error {
	head := new(big.Int).Set(from)
	limit := new(big.Int).Set(to)
	limit.Sub(limit, big.NewInt(BlockRangeMax-1))
	for {
		head.Add(head, big.NewInt(BlockRangeMax-1))
		logs, err := tf.query(from, head)
		if err != nil {
			return err
		}
		tf.logs = append(tf.logs, logs...)
		from.Add(from, big.NewInt(BlockRangeMax-1))
		if head.Cmp(to) >= 0 {
			break
		} else if from.Cmp(limit) == 1 {
			head = limit
		}
	}
	return nil
}
// query ...
func (tf *transactionsFetcher) query(from *big.Int, to *big.Int) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{
			tf.contract,
		},
	}
	return tf.client.FilterLogs(context.Background(), query)
}
// newQuery ...
func (tf *transactionsFetcher) newQuery(from *big.Int, to *big.Int) error {
	var err error
	tf.logs, err = tf.query(from, to)
	if err != nil && strings.Contains(err.Error(), "eth_getLogs block range too large") {
		return tf.newQuerySafe(from, to)
	}
	return err
}
// fetchLogs ...
func (tf *transactionsFetcher) fetchLogs(target common.Address, withTarget bool) error {
	sig, _ := tf.handler.Topic()
	for _, vLog := range tf.logs {
		switch vLog.Topics[0].Hex() {
		case sig:
			event, err := tf.handler.ToEvent(tf, &vLog)
			if err != nil {
				return err
			}
			if (withTarget && tf.handler.IsRelated(event, target) == false) || event == nil {
				continue
			}
			tf.eventLogs[vLog.TxHash] = event
		}
	}
	return nil
}
// fetchTransactions ...
func (tf *transactionsFetcher) fetchTransactions(limit uint64) error {
	fetched := uint64(0)
	for txID, _ := range tf.eventLogs {
		tx, _, err := tf.client.TransactionByHash(context.Background(), txID)
		if err != nil {
			return err
		}
		tf.transactions = append(tf.transactions, tf.handler.ToTransaction(tf, tx))
		fetched++
		if limit > 0 && fetched >= limit {
			return nil
		}
	}
	return nil
}
