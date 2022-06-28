package block

type TxOutput struct {
	Money        int64
	ScriptPubKey string // 用户名 （公钥）
}

//解锁
func (to *TxOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	return to.ScriptPubKey == address
}
