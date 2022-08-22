package handlers

import (
	"TransactionsFetcher"
)

type erc20 struct{}

var ERC20 erc20

func (erc20) Transfer() TransactionsFetcher.TransactionsFetcherHandler {
	return TransferHandler{}
}
