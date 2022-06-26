package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

type Block struct {
	//时间戳
	Timestamp int64
	//上一个区块的hash
	PrevHash []byte
	//当前区块的hash
	Hash []byte
	//交易数据
	Txs []*Transaction
	//区块高度
	Height int64
	//Nonce 值
	Nonce int64
}

//创建新的区块
func CreateBlock(txs []*Transaction, height int64, prevHash []byte) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		PrevHash:  prevHash,
		Hash:      nil,
		Txs:       txs,
		Height:    height,
		Nonce:     0,
	}
	//block.SetHash()
	//调用工作量证明方法并且返回有效的Hash和Nonce值
	pow := NewProofOfWork(block)
	block.Hash, block.Nonce = pow.Run()
	return block
}

//拼接数据，生成hash
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 2))
	height := []byte(strconv.FormatInt(b.Height, 2))
	blockBytes := bytes.Join([][]byte{b.PrevHash, timestamp, b.HashTransaction(), height}, []byte{})
	hash := sha256.Sum256(blockBytes)
	b.Hash = hash[:]
}

//生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {
	block := CreateBlock(txs, 1, []byte{32: 0})
	//block.SetHash()
	return block
}

//序列化
func (b *Block) Serialize() []byte {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(b)
	return buffer.Bytes()
}

//反序列化
func DeSerialize(data []byte) *Block {
	b := &Block{}
	encoder := gob.NewDecoder(bytes.NewReader(data))
	encoder.Decode(b)
	return b
}

//交易数据 to Hash
func (b *Block) HashTransaction() []byte {
	txHashs := [][]byte{}
	for k := range b.Txs {
		txHashs = append(txHashs, b.Txs[k].TxHash)
	}
	txHash := sha256.Sum256(bytes.Join(txHashs, []byte{}))
	return txHash[:]
}
