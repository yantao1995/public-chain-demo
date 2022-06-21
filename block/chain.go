package block

import (
	"log"

	"github.com/boltdb/bolt"
)

type BlockChain struct {
	Tip []byte //最新区块的hash
	DB  *bolt.DB
}

const (
	dbName         = "blockchain.db" //数据库名称
	blockTableName = "blocks"        //表名
	NewBlockHash   = "l"             //最新区块hash
)

//创建带有创世区块的区块链
func CreateBlockChainWithGenesisBlock() *BlockChain {

	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	blockHash := []byte{32: 0}
	err = db.Update(func(tx *bolt.Tx) error {
		//创建数据库表
		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			//创建创世区块
			genesisBlock := CreateGenesisBlock("Genesis Block")
			//存储创世区块
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//存储最新区块的hash
			err = b.Put([]byte(NewBlockHash), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			blockHash = genesisBlock.Hash
		}
		return nil
	})
	return &BlockChain{Tip: blockHash, DB: db}
}

//增加区块到区块链
func (bc *BlockChain) AddBlockToBlockChain(data string, height int64, prevHash []byte) {
	block := CreateBlock(data, height, prevHash)
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		//检查是否有数据库表
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			//存储区块
			if err := b.Put(block.Hash, block.Serialize()); err != nil {
				log.Panic(err)
			}
			//更新最新区块hash
			if err := b.Put([]byte(NewBlockHash), block.Hash); err != nil {
				log.Panic(err)
			}
		}
		return nil
	}); err != nil {
		log.Fatalln(err)
	}
	bc.Tip = block.Hash
}
