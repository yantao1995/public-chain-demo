package block

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoTableName = "utxo.db"

type UTXOSet struct {
	bc *BlockChain
}

//将所有UTXO写入数据库
func (utxo *UTXOSet) ResetUTXOSet() {
	if err := utxo.bc.DB.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(utxoTableName)); b != nil {
			if err := tx.DeleteBucket([]byte(utxoTableName)); err != nil {
				return err
			}
		}
		b, _ := tx.CreateBucket([]byte(utxoTableName))
		if b != nil {
			//[string]*Outputs
			txOutputMap := utxo.bc.FindUTXOMap()

			for keyHash, outs := range txOutputMap {
				txHash, _ := hex.DecodeString(keyHash)
				b.Put(txHash, outs.Serialize())
			}

		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

func (utxo *UTXOSet) findUTXOForAddress(address string) []*UTXO {
	var utxos []*UTXO
	utxo.bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txOutput := DeSerializeTxOutputs(v)
			for _, utxo := range txOutput.UTXOS {
				if utxo.Output.UnLockScriptPubKeyWithAddress(address) {
					utxos = append(utxos, utxo)
				}
			}
		}
		return nil
	})
	return utxos
}

//查询余额
func (utxoSet *UTXOSet) GetBalance(address string) int64 {
	UTXOS := utxoSet.findUTXOForAddress(address)
	var amount int64
	for _, utxo := range UTXOS {
		amount += utxo.Output.Money
	}
	return amount
}

//返回要凑多少钱，对应TXOutput的TX的hash和索引
func (utxoSet *UTXOSet) FindUnPackageSpendAbleUTXOs(from string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO
	spentOutput := map[string][]int{}
	//
	for _, tx := range txs {
		//txhash
		if !tx.IsCoinbaseTransaction() {
			//vins
			for _, in := range tx.Vins {
				//是否能解锁
				publicKeyHash := Base58Decoding([]byte(from))
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
			if out.UnLockScriptPubKeyWithAddress(from) {
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
	return unUTXOs
}

func (utxoSet *UTXOSet) FindSpendAbleUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	unPackageUTXOs := utxoSet.FindUnPackageSpendAbleUTXOs(from, txs)
	var money int64
	spentAbleUTXO := map[string][]int{}
	for _, utxo := range unPackageUTXOs {
		money += utxo.Output.Money
		txHash := hex.EncodeToString(utxo.TxHash)
		spentAbleUTXO[txHash] = append(spentAbleUTXO[txHash], utxo.Index)
		if money >= amount {
			return money, spentAbleUTXO
		}
	}

	//钱还不够
	utxoSet.bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeSerializeTxOutputs(v)
				for _, utxo := range txOutputs.UTXOS {
					money += utxo.Output.Money
					txHash := hex.EncodeToString(utxo.TxHash)
					spentAbleUTXO[txHash] = append(spentAbleUTXO[txHash], utxo.Index)
					if money >= amount {
						return nil
					}
				}
			}
		}
		return nil
	})

	if money < amount {
		log.Panic("余额不足")
	}

	return money, spentAbleUTXO
}

//更新
func (utxoSet *UTXOSet) Update() {

	//取最新的block
	block := utxoSet.bc.NewBlockChainIterator().Next()

	//spentUTXOMap := map[string][]int{}

	ins := []*TxInput{}

	//outMap := map[string]*TxOutputs{}

	//找到所有要删除的数据
	for _, tx := range block.Txs {
		for _, in := range tx.Vins {
			ins = append(ins, in)
		}
	}

	if err := utxoSet.bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			for _, in := range ins {
				txOutputsBytes := b.Get(in.TxHash)
				txOutputs := DeSerializeTxOutputs(txOutputsBytes)

				for _, utxo := range txOutputs.UTXOS {
					if in.Vout == utxo.Index {
						if bytes.Compare(utxo.Output.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {
							b.Delete(in.TxHash)
						}
					}
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
