package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	intVar, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Panic(err)
		return
	}

	fmt.Println("celltoHex", cellToHex(intVar))
}

func cellToHex(c int) string {
	return fmt.Sprintf("%x", []byte{byte(c / 60), byte(c % 60)})
}
