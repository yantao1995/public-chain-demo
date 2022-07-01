package main

import (
	"fmt"
	"os"
	"public-chain-demo/block"
)

func main() {
	os.Remove("blockchain.db")
	os.Remove("blockchain.db.lock")
	blockChain := block.CreateBlockChainWithGenesisBlock("1NVE728oqBcr1YMnWZ1RQADYdewxoeKuPp")
	defer blockChain.DB.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("\n___", err)
		}
	}()
	fmt.Println("\n-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println("_ balance_  a :", blockChain.GetBalance("1NVE728oqBcr1YMnWZ1RQADYdewxoeKuPp"))
	fmt.Println("-----------------------")
	blockChain.MineNewBlock([]string{"1NVE728oqBcr1YMnWZ1RQADYdewxoeKuPp"}, []string{"1BiqBw8k7hahNRFVb7WPdVF5trUZohYfvc"}, []string{"1"})
	fmt.Println("\n-----------------------")
	blockChain.Iterator()
	fmt.Println("-----------------------")
	fmt.Println("_ balance_  a :", blockChain.GetBalance("1NVE728oqBcr1YMnWZ1RQADYdewxoeKuPp"),
		"  b : ", blockChain.GetBalance("1BiqBw8k7hahNRFVb7WPdVF5trUZohYfvc"))

	// --------------------------------------------------------------

	wallets := block.NewWallets()
	wallets.Iterator()
}
