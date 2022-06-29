package main

import (
	"fmt"
	"os"
	"public-chain-demo/block"
)

func main() {
	os.Remove("blockchain.db")
	os.Remove("blockchain.db.lock")
	blockChain := block.CreateBlockChainWithGenesisBlock("a")
	defer blockChain.DB.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("\n___", err)
		}
	}()
	fmt.Println("\n-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println("_ balance_  a :", blockChain.GetBalance("a"),
		"  b : ", blockChain.GetBalance("b"),
		"  c : ", blockChain.GetBalance("c"))
	fmt.Println("-----------------------")
	fmt.Println("-----------------------")
	blockChain.MineNewBlock([]string{"a", "b"}, []string{"b", "c"}, []string{"3", "2"})
	blockChain.MineNewBlock([]string{"a"}, []string{"c"}, []string{"1"})
	fmt.Println("\n-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println("_ balance_  a :", blockChain.GetBalance("a"),
		"  b : ", blockChain.GetBalance("b"),
		"  c : ", blockChain.GetBalance("c"))

}
