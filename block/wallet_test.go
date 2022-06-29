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
