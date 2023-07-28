package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	ethUnit "github.com/DeOne4eg/eth-unit-converter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/scylladb/termtables"
)

var (
	wss = fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", os.Getenv("INFURA_SECRET"))
	url = fmt.Sprintf("https://mainnet.infura.io/v3/%s", os.Getenv("INFURA_SECRET"))
)

func watch() {
	backend, err := ethclient.Dial(url)
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}

	rpcCli, err := rpc.Dial(wss)
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}
	gcli := gethclient.New(rpcCli)

	txch := make(chan common.Hash, 100)
	_, err = gcli.SubscribePendingTransactions(context.Background(), txch)

	if err != nil {
		log.Printf("failed to SubscribePendingTransactions: %v", err)
		return
	}

	for {
		select {
		case txhash := <-txch:
			tx, _, err := backend.TransactionByHash(context.Background(), txhash)
			if err != nil {
				continue
			}
			if tx.To() == nil {
				r, err := backend.TransactionReceipt(context.Background(), txhash)
				if err != nil {
					continue
				}
				if strings.ToLower(r.ContractAddress.Hex()) != "" {
					gasFeeCap, _ := backend.SuggestGasPrice(context.Background()) // maxFeePerGas 获取当前Gas的价格
					gasPriceGwei := ethUnit.NewWei(gasFeeCap).GWei()
					t := termtables.CreateTable()
					t.AddHeaders("交易Hash", "合约地址", "建议Gas")
					t.AddRow(tx.Hash().Hex(), strings.ToLower(r.ContractAddress.Hex()), fmt.Sprintf("%2.f GWei", gasPriceGwei))
					fmt.Println(t.Render())
				}
			}
		}
	}
}

func main() {
	go watch()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
