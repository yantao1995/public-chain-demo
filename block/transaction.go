package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

//UTXO
type Transaction struct {
	// 交易hash
	TxHash []byte
	//输入
	Vins []*TxInput
	//输出
	Vouts []*TxOutput
}

//两种情况

//创世区块创建的 Transaction
func NewCoinBaseTransaction(address string) *Transaction {
	//消费
	txInput := &TxInput{[]byte{}, -1, "Genesis Data"}
	//未消费
	txOutput := &TxOutput{10, address}

	txCoinbase := &Transaction{
		TxHash: []byte{},
		Vins:   []*TxInput{txInput},
		Vouts:  []*TxOutput{txOutput},
	}
	//设置hash
	txCoinbase.HashTransaction()
	return txCoinbase
}

//设置hash值
func (tx *Transaction) HashTransaction() {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(tx)
	hash := sha256.Sum256(buffer.Bytes())
	tx.TxHash = hash[:]
}

//转账时产生的 Transaction
