package block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
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
	fmt.Println("正常迭代区块....")
	defer fmt.Println("区块迭代完成.")
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
				fmt.Println("from", hex.EncodeToString(block.Txs[k].Vins[0].PublicKey), "to", block.Txs[k].Vouts[0].Ripemd160Hash, "amount ", block.Txs[k].Vouts[0].Money)
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
		tx := NewSimpleTransaction(from[index], to[index], val, bc, txs)
		txs = append(txs, tx)
	}
	//奖励
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)

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

	//未打包的
	_txs := []*Transaction{}

	//在建立新区块之前对 txs 进行签名验证
	for _, tx := range txs {
		if !bc.VerifyTransaction(tx, _txs) {
			log.Panic("签名验证失败...")
		}
		_txs = append(_txs, tx)
	}

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
				publicKeyHash := Base58Decoding([]byte(address))
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-addressCheckSumLen]
				if in.UnLockWith160Hash(ripemd160Hash) {
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
				}
				utxo := &UTXO{tx.TxHash, index, out}
				unUTXOs = append(unUTXOs, utxo)
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

			for i := len(block.Txs) - 1; i >= 0; i-- {
				handle(block.Txs[i])
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
		panic(fmt.Sprintf("%s 余额不足\n", from))
	}
	return value, spendAbleUTXO
}

func (bc *BlockChain) SignTransaction(tx *Transaction, private ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vins {
		prevTX, err := bc.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	tx.Sign(private, prevTXs)
}

func (bc *BlockChain) FindTransaction(ID []byte, txs []*Transaction) (Transaction, error) {

	for _, tx := range txs {
		if bytes.Compare(tx.TxHash, ID) == 0 {
			return *tx, nil
		}
	}

	var block *Block
	currentHash := bc.Tip
	var trx *Transaction
	for trx == nil {
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
				if bytes.Compare(tx.TxHash, ID) == 0 {
					trx = tx
					return nil
				}
			}
			//
			currentHash = block.PrevHash
			return nil
		}); err != nil {
			break
		}
	}
	if trx == nil {
		return Transaction{}, nil
	}
	return *trx, nil
}

//验证数字签名
func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vins {
		prevTX, err := bc.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) FindUTXOMap() map[string]*TxOutputs {
	var block *Block
	currentHash := bc.Tip
	utxoMaps := map[string]*TxOutputs{}
	spentAbleUTXOMap := make(map[string][]*TxInput)
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

			for i := len(block.Txs) - 1; i >= 0; i-- {

				txOutputs := &TxOutputs{[]*UTXO{}}
				tx := block.Txs[i]
				txHash := hex.EncodeToString(tx.TxHash)
				if !tx.IsCoinbaseTransaction() {
					for _, txInput := range tx.Vins {
						txHash := hex.EncodeToString(txInput.TxHash)
						spentAbleUTXOMap[txHash] = append(spentAbleUTXOMap[txHash], txInput)
					}
				}

			outLab:
				for outIndex, out := range tx.Vouts {
					if txInputs, ok := spentAbleUTXOMap[txHash]; ok {
						isSpent := false
						for _, in := range txInputs {
							outPublicKey := out.Ripemd160Hash
							inPublicKey := in.PublicKey
							if bytes.Compare(outPublicKey, Ripemd160Hash(inPublicKey)) == 0 {
								if outIndex == in.Vout {
									isSpent = true
									continue outLab
								}
							}
						}
						if !isSpent {
							utxo := &UTXO{tx.TxHash, outIndex, out}
							txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
						}
					} else {
						utxo := &UTXO{tx.TxHash, outIndex, out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}
				}

				utxoMaps[txHash] = txOutputs
			}

			//
			currentHash = block.PrevHash
			return nil
		}); err != nil {
			break
		}
	}
	return utxoMaps
}
