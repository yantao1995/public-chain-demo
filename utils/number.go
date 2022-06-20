package utils

import "fmt"

func IntToHex(num int64) []byte {
	return []byte(fmt.Sprintf("%x", num))
}
