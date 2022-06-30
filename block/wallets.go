package block

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	walletFile = "wallets.dat"
)

type Wallets struct {
	WalletsMap map[string]*Wallet //地址-钱包
}

//创建钱包集合
func NewWallets() *Wallets {
	//取出文件内的钱包
	w := &Wallets{
		WalletsMap: map[string]*Wallet{},
	}
	if err := w.LoadFile(); err != nil {
		fmt.Println("加载钱包文件异常:", err)
	}
	return w
}

//加载钱包文件
func (w *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}
	gob.Register(elliptic.P256())
	return gob.NewDecoder(bytes.NewReader(fileContent)).Decode(w)
}

//创建新钱包
func (w *Wallets) CreateNewWallet() {
	wallet := NewWallet()
	fmt.Printf("Address : %s\n", wallet.GetAddress())
	w.WalletsMap[string(wallet.GetAddress())] = wallet
	w.SaveWallets()
}

//存储到文件
func (w *Wallets) SaveWallets() {
	buffer := bytes.Buffer{}
	//注册为了序列化任何类型
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(w); err != nil {
		log.Panic(err)
	}
	//覆盖
	if err := ioutil.WriteFile(walletFile, buffer.Bytes(), 0644); err != nil {
		log.Panic(err)
	}
}

//输出所有钱包地址
func (w *Wallets) Iterator() {
	for k := range w.WalletsMap {
		fmt.Println(k)
	}
}
