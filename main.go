package main

import (
	"fmt"
	"os"
	"public-chain-demo/block"
)

func main() {
	os.Remove("blockchain.db")
	blockChain := block.CreateBlockChainWithGenesisBlock("address")
	defer blockChain.DB.Close()
	blockChain.MineNewBlock([]string{"a"}, []string{"b"}, []string{"1"})
	fmt.Println("-----------------------")
	blockChain.Iterator()
}
