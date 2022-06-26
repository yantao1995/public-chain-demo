package block

import (
	"errors"
	"fmt"
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
func CreateBlockChainWithGenesisBlock(address string) *BlockChain {

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
			//创建 Coinbase Transaction
			txCoinbase := NewCoinBaseTransaction(address)
			//创建创世区块
			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})
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
	block := CreateBlock(nil, height, prevHash)
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

//区块迭代
func (bc *BlockChain) Iterator() {
	var block *Block
	currentHash := bc.Tip
	for {
		if err := bc.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blockTableName))
			if b == nil {
				return errors.New("EOF")
			}
			data := b.Get(currentHash)
			if data == nil {
				return errors.New("EOF")
			}
			block = DeSerialize(data)
			fmt.Println("block: ", string(block.HashTransaction()), "currentHash: ", fmt.Sprintf("%x", currentHash))
			currentHash = block.PrevHash
			return nil
		}); err != nil {
			break
		}
	}
}

//挖矿
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction

	var block *Block

	bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			hash := bucket.Get([]byte(NewBlockHash))
			blockBytes := bucket.Get(hash)
			block = DeSerialize(blockBytes)
		}
		return nil
	})

	//创建新区块
	block = CreateBlock(txs, block.Height+1, block.Hash)

	//将新区块存储到数据库
	bc.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			bucket.Put(block.Hash, block.Serialize())
			bucket.Put([]byte(NewBlockHash), block.Hash)
			bc.Tip = block.Hash
		}
		return nil
	})
}
