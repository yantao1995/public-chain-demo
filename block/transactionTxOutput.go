package block

import "bytes"

type TxOutput struct {
	Money         int64
	Ripemd160Hash []byte //公钥的hash  与公钥是对应的
}

//上锁
func (to *TxOutput) Lock(address string) {
	//25字节
	publicKeyHash := Base58Decoding([]byte(address))
	//获取 除版本号 和 checksum 的值
	to.Ripemd160Hash = publicKeyHash[1 : len(publicKeyHash)-addressCheckSumLen]
}

//解锁
func (to *TxOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	publicKeyHash := Base58Decoding([]byte(address))
	hash160 := publicKeyHash[1 : len(to.Ripemd160Hash)-addressCheckSumLen]
	return bytes.Compare(hash160, to.Ripemd160Hash) == 0
}

//相当于给钱上锁
func NewTXOutput(value int64, address string) *TxOutput {
	txOutput := &TxOutput{value, nil}
	//设置ripemd160 hash
	txOutput.Lock(address)
	return txOutput
}
