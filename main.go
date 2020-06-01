package main

import (
	"os"
	"strconv"
)

func main() {
	arguments := os.Args[1:]
	if arguments[0] == "-c" {
		ngramSize, _ := strconv.Atoi(arguments[1])
		fileName := arguments[2]
		compress(fileName, ngramSize)
	} else if arguments[0] == "-d" {
		fileName := arguments[1]
		decompress(fileName)
	}
}
