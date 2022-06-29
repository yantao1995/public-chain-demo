package block

import (
	"bytes"
	"math/big"
	"public-chain-demo/utils"
)

//与 base64相比去掉了6个容易混淆的，去掉0，大写的O、大写的I、小写的L、/、+/、+影响双击选择
var base58 = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

//Base58编码
func Base58Encoding(strByte []byte) []byte {
	// 转换十进制
	strTen := big.NewInt(0).SetBytes(strByte)
	// 取出余数
	var modSlice []byte
	for strTen.Cmp(big.NewInt(0)) > 0 {
		mod := big.NewInt(0) //余数
		strTen58 := big.NewInt(58)
		strTen.DivMod(strTen, strTen58, mod)             //取余运算
		modSlice = append(modSlice, base58[mod.Int64()]) //存储余数,并将对应值放入其中
	}
	// 处理0就是1的情况 0使用字节'1'代替
	for _, elem := range strByte {
		if elem != 0 {
			break
		} else if elem == 0 {
			modSlice = append(modSlice, byte('1'))
		}
	}
	utils.ReverseBytes(modSlice)
	return modSlice
}

//Base58解码
func Base58Decoding(strByte []byte) []byte { //Base58解码
	ret := big.NewInt(0)
	for _, byteElem := range strByte {
		index := bytes.IndexByte(base58, byteElem) //获取base58对应数组的下标
		ret.Mul(ret, big.NewInt(58))               //相乘回去
		ret.Add(ret, big.NewInt(int64(index)))     //相加
	}
	return ret.Bytes()
}
