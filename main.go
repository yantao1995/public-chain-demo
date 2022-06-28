package main

import (
	"fmt"
	"os"
	"public-chain-demo/block"
)

func main() {
	os.Remove("blockchain.db")
	blockChain := block.CreateBlockChainWithGenesisBlock("a")
	defer blockChain.DB.Close()
	fmt.Println("\n-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println(blockChain.GetBalance("a"))
	fmt.Println("-----------------------")
	fmt.Println(blockChain.UnUTXOs("a", nil))
	fmt.Println("-----------------------")
	blockChain.MineNewBlock([]string{"a"}, []string{"b"}, []string{"1"})
	fmt.Println("-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println(blockChain.UnUTXOs("b", nil))
	fmt.Println(blockChain.GetBalance("a"))
}
