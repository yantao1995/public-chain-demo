package block

import (
	"fmt"
	"testing"
)

func TestWallet(t *testing.T) {
	wallet := NewWallet()
	address := wallet.GetAddress()
	isValid := wallet.IsValidAddress(address)
	fmt.Printf("%s %v\n", address, isValid)
}

func TestWallets(t *testing.T) {
	wallets := NewWallets()
	fmt.Println(wallets.WalletsMap)
	Wallet := NewWallet()
	wallets.WalletsMap[string(Wallet.GetAddress())] = Wallet
	fmt.Println(wallets.WalletsMap)
	wallets.CreateNewWallet()
}
