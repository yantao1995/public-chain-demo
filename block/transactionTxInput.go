package block

import "bytes"

type TxInput struct {
	//交易hash
	TxHash []byte
	//在UTXO模型的未花费区块的，TxOutput在Vout里面的索引
	Vout      int
	Signature []byte //数字签名
	PublicKey []byte //公钥
}

//判断当前消费是否为address的钱
func (ti *TxInput) UnLockWith160Hash(ripemd160Hash []byte) bool {
	publicKey := Ripemd160Hash(ti.PublicKey)
	return bytes.Compare(publicKey, ripemd160Hash) == 0
}
