package block

type BlockChain struct {
	Blocks []*Block //存储有序区块
}

//创建带有创世区块的区块链
func CreateBlockChainWithGenesisBlock() *BlockChain {
	genesisBlock := CreateGenesisBlock("Genesis Block")
	return &BlockChain{[]*Block{genesisBlock}}
}

//增加区块到区块链
func (bc *BlockChain) AddBlockToBlockChain(data string, height int64, prevHash []byte) {
	block := CreateBlock(data, height, prevHash)
	bc.Blocks = append(bc.Blocks, block)
}
