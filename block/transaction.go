package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
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
	txInput := &TxInput{[]byte{}, -1, nil, []byte{}}
	//未消费
	txOutput := NewTXOutput(10, address)

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

	wallets := NewWallets()
	wallet := wallets.WalletsMap[from]

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
			txInput := &TxInput{txHashBytes, index, nil, wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = NewTXOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	//设置hash值
	tx.HashTransaction()

	//进行签名
	bc.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx
}

//判断当前交易是否为coinbase交易  false为正常的区块内交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("没有找到对应的Transaction")
		}
	}
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vins {
		prevTXs := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTXs.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil
		//签名
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vins[inID].Signature = signature
	}

}

//拷贝一份新的用于签名
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput
	for _, vin := range tx.Vins {
		inputs = append(inputs, &TxInput{TxHash: vin.TxHash, Vout: vin.Vout})
	}

	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TxOutput{
			Money:         vout.Money,
			Ripemd160Hash: vout.Ripemd160Hash,
		})
	}

	txCopy := Transaction{tx.TxHash, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Hash() []byte {
	txCopy := *tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx Transaction) Serialize() []byte {
	buffer := bytes.Buffer{}
	if err := gob.NewEncoder(&buffer).Encode(tx); err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//验证签名
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("Transaction 异常 ")
		}
	}

	txCopy := tx.TrimmedCopy()

	curve := elliptic.P256()

	for inID, vin := range txCopy.Vins {
		prevTX := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTX.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		//私钥id
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) {
			return false
		}
	}
	return true
}
