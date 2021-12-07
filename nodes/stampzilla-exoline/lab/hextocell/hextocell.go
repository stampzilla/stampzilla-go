package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("hexToCell", hexToCell(os.Args[1]))
}
func hexToCell(s string) int {
	v, err := hex.DecodeString(s)
	if err != nil {
		log.Println(err)
		return 0
	}
	first := int(v[0])
	second := int(v[1])
	return first*60 + second
}
