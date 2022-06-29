package utils

import "fmt"

func IntToHex(num int64) []byte {
	return []byte(fmt.Sprintf("%x", num))
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
