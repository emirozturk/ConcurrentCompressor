package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	arguments := os.Args[1:]
	if arguments[0] == "-c" {
		if len(arguments)==3 {
			fileName := arguments[1]
			ngramSize, _ := strconv.Atoi(arguments[2])
			stream := compress(readFile(fileName), ngramSize)
			writeCCStream(fileName,stream)
		}else{
			fmt.Print("Usage: CC -c fileName ngramLength")
		}
	} else if arguments[0] == "-d" {
		if len(arguments)==2{
			fileName := arguments[1]
			output := decompress(readCCStream(fileName))
			writeFile(fileName,output)
		}else{
			fmt.Print("Usage: CC -d fileName")
		}
	}
}
