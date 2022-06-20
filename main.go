package main

import (
	"fmt"
	"public-chain-demo/block"
)

func main() {
	chain := block.CreateBlockChainWithGenesisBlock()
	chain.AddBlockToBlockChain("Send 1 BTC to Bob", 1, []byte{})
	fmt.Println()
	fmt.Println(len([]rune("00fd83f23d3245d13e663f99bbb5316f23f7135b97901658a67e6498ddd2a35f")))
}
