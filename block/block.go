package block

import (
	"bytes"
	"crypto/sha256"
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
	Data []byte
	//区块高度
	Height int64
	//Nonce 值
	Nonce int64
}

//创建新的区块
func CreateBlock(data string, height int64, prevHash []byte) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		PrevHash:  prevHash,
		Hash:      nil,
		Data:      []byte(data),
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
	blockBytes := bytes.Join([][]byte{b.PrevHash, timestamp, b.Data, height}, []byte{})
	hash := sha256.Sum256(blockBytes)
	b.Hash = hash[:]
}

//生成创世区块
func CreateGenesisBlock(data string) *Block {
	block := CreateBlock(data, 1, []byte{32: 0})
	block.SetHash()
	return block
}
