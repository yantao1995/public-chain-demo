package block

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"public-chain-demo/utils"
)

// 0000 0000 0000 0000 1000 0001 1010 ... 0001
// 256 位hash里面至少要有16个0
const targetBit = 8

type ProofOfWork struct {
	Block  *Block   //要验证的区块
	target *big.Int //大数据存储  (难度值  n表示hash前面有n个0)
}

//创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {
	// big.Int 对象
	target := big.NewInt(1)
	//左移256-targetBit位
	target = target.Lsh(target, 256-targetBit)
	////相当于难度已经定了
	return &ProofOfWork{block, target}
}

func (p *ProofOfWork) Run() ([]byte, int64) {
	nonce := int64(0)
	var hashInt *big.Int = &big.Int{} //存储新生成的hash
	hash := [32]byte{}
	for ; ; nonce++ {
		//将block属性拼接成字节数组
		dataBytes := bytes.Join([][]byte{
			p.Block.PrevHash, utils.IntToHex(p.Block.Timestamp),
			p.Block.Data, utils.IntToHex(targetBit), utils.IntToHex(nonce)}, []byte{})
		//生成hash
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x", hash)
		//将hash存储到 hashInt
		hashInt.SetBytes(hash[:])
		//判断hashInt 是否小于 target //判断hash的有效性,如果满足，跳出循环
		if p.target.Cmp(hashInt) == 1 {
			break
		}
	}
	return hash[:], nonce
}

//验证区块的hash值是否有效
func (p *ProofOfWork) VerifyProofOfWork(block *Block) bool {
	var hashInt big.Int
	hashInt.SetBytes(p.Block.Hash)
	if p.target.Cmp(&hashInt) == 1 {
		return true
	}
	return false
}
