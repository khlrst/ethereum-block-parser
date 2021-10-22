package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"

	web3 "github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

type Transaction struct {
	Hash        web3.Hash
	From        web3.Address
	Input       string
	Value       big.Int
	Nonce       uint64
	BlockHash   web3.Hash
	BlockNumber uint64
}

func ExtractOpenseaTransactions(input *web3.Block, transactions *[]Transaction) {
	openseaAddress := web3.HexToAddress("0x7Be8076f4EA4A4AD08075C2508e481d6C946D12b")
	for i := 0; i < len(input.Transactions); i++ {
		if input.Transactions[i].To != nil && len(input.Transactions[i].Input) > 0 {
			if *input.Transactions[i].To == openseaAddress {
				selector := hex.EncodeToString(input.Transactions[i].Input[0:4])
				if selector == "ab834bab" { //AtomicMatch selector
					// *transactions = append(*transactions, Transaction{
					// 	Hash:        input.Transactions[i].Hash,
					// 	From:        input.Transactions[i].From,
					// 	Input:       hex.EncodeToString(input.Transactions[i].Input),
					// 	Value:       *input.Transactions[i].Value,
					// 	Nonce:       input.Transactions[i].Nonce,
					// 	BlockHash:   input.Transactions[i].BlockHash,
					// 	BlockNumber: input.Transactions[i].BlockNumber,
					// })
					fmt.Println("buy order static target: ", hex.EncodeToString(input.Transactions[i].Input)[328:392][24:64], "  sell order static target: ", hex.EncodeToString(input.Transactions[i].Input)[776:840][24:64])
				}
			}
		}

	}
}

func fetchBlocks(start uint64, end uint64, client *jsonrpc.Client, blocks chan *web3.Block) {
	for end >= start {
		block, err := client.Eth().GetBlockByNumber(web3.BlockNumber(start), true)
		if err != nil {
			panic(err)
		}
		blocks <- block
		start++
	}
}

func main() { // supply infura API key, depth of blocks
	var infuraApiKey string
	var depth uint64

	flag.StringVar(&infuraApiKey, "i", infuraApiKey, "Specify infuraApiKey. Cannot be null")
	flag.Uint64Var(&depth, "d", depth, "Specify depth. Cannot be 0")
	// read args
	flag.Parse()
	// get a client
	client, err := jsonrpc.NewClient(fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraApiKey))
	if err != nil {
		panic(err)
	}
	// get depth and last block number
	end, err := client.Eth().BlockNumber()
	if err != nil {
		panic(err)
	}

	start := end - depth + 1 // escape +1 block fetch
	//create channel for node requests queue
	blocks := make(chan *web3.Block, 1)
	// create transaction storage slice
	transactions := make([]Transaction, 0)
	go func() {
		for {
			block, more := <-blocks
			if more {
				go ExtractOpenseaTransactions(block, &transactions)
			} else {
				return
			}
		}
	}()
	fetchBlocks(start, end, client, blocks)
	close(blocks)
	// for _, transaction := range transactions {
	// 	fmt.Println("buy order static target: ", transaction.Input[328:392][24:64], "  sell order static target: ", transaction.Input[776:840][24:64]) // print arguments from atomic match
	// }
}
