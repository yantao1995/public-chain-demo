package main

import (
	"fmt"
	"os"
	"public-chain-demo/block"
)

func main() {
	os.Remove("blockchain.db")
	blockChain := block.CreateBlockChainWithGenesisBlock()
	defer blockChain.DB.Close()
	blockChain.AddBlockToBlockChain("Send 1RMB to zhangsan", 2, blockChain.Tip)
	fmt.Println()
	blockChain.AddBlockToBlockChain("Send 2RMB to lisi", 3, blockChain.Tip)
	fmt.Println()
	fmt.Println("-----------------------")
	blockChain.Iterator()
}
