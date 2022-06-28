package block

type TxInput struct {
	//交易hash
	TxHash []byte
	//在UTXO模型的未花费区块的，TxOutput在Vout里面的索引
	Vout int
	//用户名
	ScriptSig string //用户签名
}

//判断当前消费是否为address的钱
func (ti *TxInput) UnLockWithAddress(address string) bool {
	return ti.ScriptSig == address
}
