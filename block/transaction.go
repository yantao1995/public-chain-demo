package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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
func NewSimpleTransaction(from, to string, amount int, bc *BlockChain, txs []*Transaction) *Transaction {

	//unUTXOs := bc.UnSpentTransactionsWithAddress(from)

	money, spendAbleUTXODict := bc.FindSpendAbleUTXOs(from, amount, txs)

	var (
		txInputs  []*TxInput
		txOutputs []*TxOutput
	)

	//消费
	for txHash, indexArray := range spendAbleUTXODict {
		txHashBytes, _ := hex.DecodeString(txHash)
		//transaction vout 的索引
		for _, index := range indexArray {
			txInput := &TxInput{txHashBytes, index, from}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := &TxOutput{int64(amount), to}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TxOutput{int64(money) - int64(amount), from}
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.HashTransaction()
	return tx
}

//判断当前交易是否为coinbase交易  false为正常的区块内交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}
