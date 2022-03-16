package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

type Output struct {
	Hash     web3.Hash `json: "hash"`
	RootBuy  string    `json: "rootBuy"`
	RootSell string    `json: "rootSell"`
}

func ExtractOpenseaTransactions(input *web3.Block, transactions *[]Transaction, outputs *[]Output) {
	openseaAddress := web3.HexToAddress("0x7f268357A8c2552623316e2562D90e642bB538E5")

	for i := 0; i < len(input.Transactions); i++ {
		if input.Transactions[i].To != nil && len(input.Transactions[i].Input) > 0 {
			if *input.Transactions[i].To == openseaAddress {
				selector := hex.EncodeToString(input.Transactions[i].Input[0:4])
				if selector == "ab834bab" && hex.EncodeToString(input.Transactions[i].Input)[3464:3472] == "fb16a595" { //AtomicMatch selector and merkle validator selector
					*outputs = append(*outputs, Output{
						Hash:     input.Transactions[i].Hash,
						RootBuy:  hex.EncodeToString(input.Transactions[i].Input)[3728:3792],
						RootSell: hex.EncodeToString(input.Transactions[i].Input)[4304:4368],
					})
					fmt.Println(input.Transactions[i].Hash, " buy root hash: ", hex.EncodeToString(input.Transactions[i].Input)[3728:3792], "  sell root hash: ", hex.EncodeToString(input.Transactions[i].Input)[4304:4368])
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
	// create array of outputs
	outputs := make([]Output, 0)
	go func() {
		for {
			block, more := <-blocks
			if more {
				go ExtractOpenseaTransactions(block, &transactions, &outputs)
			} else {
				return
			}
		}
	}()
	fetchBlocks(start, end, client, blocks)
	close(blocks)
	blob, _ := json.Marshal(outputs)
	ioutil.WriteFile("output.json", blob, 0644)
}
