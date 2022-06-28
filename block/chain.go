package block

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

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
			fmt.Println("当前区块高度：", block.Height, " hash: ", fmt.Sprintf("%x", currentHash), "\n交易数据:")
			for k := range block.Txs {
				fmt.Println("from", block.Txs[k].Vins[0].ScriptSig, "to", block.Txs[k].Vouts[0].ScriptPubKey, "amount ", block.Txs[k].Vouts[0].Money)
			}
			currentHash = block.PrevHash
			return nil
		}); err != nil {
			break
		}
	}
}

//挖矿
func (bc *BlockChain) MineNewBlock(from, to, amount []string) {
	//建立新交易
	var txs []*Transaction
	for index := range from {
		val, _ := strconv.Atoi(amount[index])
		tx := NewSimpleTransaction(from[index], to[index], val, bc, nil)
		txs = append(txs, tx)
	}
	//

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

//TxOutput未花费的对应地址列表
func (bc *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	var block *Block
	currentHash := bc.Tip

	//
	var unUTXOs []*UTXO
	spentOutput := map[string][]int{}
	//

	handle := func(tx *Transaction) {
		//txhash
		if !tx.IsCoinbaseTransaction() {
			//vins
			for _, in := range tx.Vins {
				//是否能解锁
				if in.UnLockWithAddress(address) {
					key := hex.EncodeToString(in.TxHash)
					spentOutput[key] = append(spentOutput[key], in.Vout)
				}
			}
		}
		//vouts
	outLab:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				if indexArrays, ok := spentOutput[hex.EncodeToString(tx.TxHash)]; ok {
					for _, indexArray := range indexArrays {
						if index == indexArray {
							continue outLab
						}
					}
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}
			}
		}
	}

	//处理本区块中未打包的交易
	for _, tx := range txs {
		handle(tx)
	}

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
			//

			for _, tx := range block.Txs {
				handle(tx)
			}
			//
			currentHash = block.PrevHash
			return nil
		}); err != nil {
			break
		}
	}
	return unUTXOs
}

//查询某地址余额
func (bc *BlockChain) GetBalance(address string) int64 {
	utxos := bc.UnUTXOs(address, nil)

	var amount int64
	for _, out := range utxos {
		amount += out.Output.Money
	}
	return amount
}

//转账时查找可用的UTXO
func (bc *BlockChain) FindSpendAbleUTXOs(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	//获取所有的UXTO
	utxos := bc.UnUTXOs(from, txs)
	//遍历 utxos
	var value int64
	spendAbleUTXO := map[string][]int{}
	for _, utxo := range utxos {
		value += int64(utxo.Output.Money)
		hash := hex.EncodeToString(utxo.TxHash)
		spendAbleUTXO[hash] = append(spendAbleUTXO[hash], utxo.Index)
		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Printf("%s 余额不足\n", from)
		os.Exit(1)
	}
	return value, spendAbleUTXO
}
