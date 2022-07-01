package block

import (
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

	return nil
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
