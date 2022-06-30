package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressCheckSumLen = 4
)

type Wallet struct {
	//私钥
	PrivateKey ecdsa.PrivateKey
	//公钥
	PublicKey []byte
}

//创建钱包
func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()
	return &Wallet{privateKey, publicKey}
}

//通过私钥产生公钥  椭圆曲线
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	pubKey := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pubKey
}

//判断地址有效
func (w *Wallet) IsValidAddress(address []byte) bool {
	versionPublicCheckSumBytes := Base58Decoding(address)
	bts := versionPublicCheckSumBytes[len(versionPublicCheckSumBytes)-addressCheckSumLen:]
	versionRipemd160 := versionPublicCheckSumBytes[:len(versionPublicCheckSumBytes)-addressCheckSumLen]
	//base58编码将首位版本 0x00 去掉了。所以需要手动补0x00， 如果非 0x00，则没有问题
	if version == 0x00 {
		versionRipemd160 = append([]byte{version}, versionRipemd160...)
	}
	checkBytes := CheckSum(versionRipemd160)
	return bytes.Compare(checkBytes, bts) == 0
}

//根据公钥获取地址
func (w *Wallet) GetAddress() []byte {
	// hash160
	ripemd160Hash := Ripemd160Hash(w.PublicKey)
	versionRipemd160Hash := append([]byte{version}, ripemd160Hash...)
	checkSumBytes := CheckSum(versionRipemd160Hash)
	bytes := append(versionRipemd160Hash, checkSumBytes...)
	return Base58Encoding(bytes)
}

//取两次sha256之后的前4个字节
func CheckSum(payload []byte) []byte {
	firstSha256 := sha256.Sum256(payload)
	secondSha256 := sha256.Sum256(firstSha256[:])
	return secondSha256[:addressCheckSumLen]
}

//获取ripemd160 值
func Ripemd160Hash(publicKey []byte) []byte {
	//256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)
	//160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)
	return ripemd160.Sum(nil)
}
