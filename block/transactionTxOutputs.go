package block

import (
	"bytes"
	"encoding/gob"
)

type TxOutputs struct {
	UTXOS []*UTXO
}

//序列化
func (b *TxOutputs) Serialize() []byte {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(b)
	return buffer.Bytes()
}

//反序列化
func DeSerializeTxOutputs(data []byte) *TxOutputs {
	b := &TxOutputs{}
	encoder := gob.NewDecoder(bytes.NewReader(data))
	encoder.Decode(b)
	return b
}
