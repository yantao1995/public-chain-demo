package main

import (
	"fmt"
	"math/big"
	"public-chain-demo/block"
)

func main() {
	blockChain := block.CreateBlockChainWithGenesisBlock()
	defer blockChain.DB.Close()
	hashInt := big.Int{}
	hashInt.SetBytes(blockChain.Tip)
	str := [64]byte{}
	txt := hashInt.Text(16)
	for i := 31; i >= 0; i-- {
		str[i] = txt[i]
	}
	fmt.Println("\n", string(str[:]))
	fmt.Println("\n", hashInt.Text(16))
	blockChain.AddBlockToBlockChain("Send 100RMB to wangqiang", 2, blockChain.Tip)
	fmt.Printf("\n%x\n", string(blockChain.Tip))
	blockChain.AddBlockToBlockChain("Send 100RMB to wangqiang", 3, blockChain.Tip)
	fmt.Printf("\n%x\n", blockChain.Tip)
}
