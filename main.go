package main

import (
	"flag"
	"io/ioutil"
	"strconv"
)
func SecondPass(input []byte){

}

func FirstPass(input []byte){

}

func Compress(fileName string,nGramSize int){
	input,_ := ioutil.ReadFile(fileName)
	FirstPass(input)
	SecondPass(input)
}

func Decompress(fileName string){

}

func main() {
	var fileName string
	var mode string
	var nString string
	flag.StringVar(&fileName,"file","","Path of the file to be compressed/decompressed")
	flag.StringVar(&mode,"mode","c","c for compression and d for decompression")
	flag.StringVar(&nString,"n","3","n-gram length")
	flag.Parse()
	n,_ := strconv.Atoi(nString)
	if n <= 0 {
		n = 3
	}
	if fileName!=""{
		if mode == "c" {
			Compress(fileName,n)
		} else if mode == "d" {
			Decompress(fileName)
		}
	}
}
