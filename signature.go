package TransactionsFetcher

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type erc20 struct{}

var ERC20 erc20

func (erc20) Transfer() (string, string) {
	return crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex(), "Transfer"
}

func logToTransfer(log types.Log) interface{} {
	var transferEvent LogTransfer
	// if err := tf.contractABI.UnpackIntoInterface(&transferEvent, topicName, log.Data); err != nil {
	// 	return err
	//}
	transferEvent.From = common.HexToAddress(log.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(log.Topics[2].Hex())
	return transferEvent
}
