package main

import (
	"encoding/binary"
)

var D1 [256][]byte
var D2 [65536][]byte
var ngramSize int

func createOutput(bvBytes []byte,streams [3][]byte) []byte{
	bitCount := len(bvBytes[1:])*8
	output := make([]byte,bitCount*4/2)
	redundantBits:=bvBytes[0]
	length:=bitCount-int(redundantBits)

	mask := [4]byte{192,48,12,3}
	counter := [3]int{}
	var outputCounter,bvCounter,shiftCounter int
	bvCounter=0
	for i:=0;i<length;i+=2{
		shiftCounter = (i>>1) & 3
		if shiftCounter == 0 {
			bvCounter++
		}
		bitResult := (bvBytes[bvCounter] & mask[shiftCounter]) >> (6 - shiftCounter*2)
		copy(output[outputCounter:],streams[bitResult][counter[bitResult]*ngramSize:(counter[bitResult]+1)*ngramSize])
		counter[bitResult]++
		outputCounter+=ngramSize
	}
	return output
}

func createDictionaryArray(stream []byte,index int){
	counter := 0
	if index == 1 {
		for i:=1;i<len(stream);i+=ngramSize{
			D1[counter] = stream[i:i+ngramSize]
			counter++
		}
	} else if index == 2 {
		for i:=0;i<len(stream);i+=ngramSize{
			D2[counter] = stream[i:i+ngramSize]
			counter++
		}
	}
}
func createArray(channel chan concurrentStream,stream []byte,index int){
	cc := concurrentStream{}
	cc.id = index
	counter:=0
	if index == 1{
		cc.stream = make([]byte,len(stream)*ngramSize)
		for i:=0;i<len(stream);i++{
			copy(cc.stream[counter:],D1[stream[i]])
			counter+=ngramSize
		}
	}else if index ==2{
		cc.stream = make([]byte,len(stream)*ngramSize/2)
		for i:=0;i<len(stream);i+=2{
			copy(cc.stream[counter:],D2[binary.BigEndian.Uint16(stream[i:i+2])])
			counter+=ngramSize
		}
	}else if index == 3{
		cc.stream = make([]byte,len(stream))
		for i:=0;i<len(stream);i+=ngramSize{
			copy(cc.stream[counter:],stream[i:i+ngramSize])
			counter+=ngramSize
		}
	}

	channel <-cc
}
func decompress(stream ccStream) []byte {
	ngramSize =int(stream.D1[0])
	createDictionaryArray(stream.D1,1)
	createDictionaryArray(stream.D2,2)

	channel := make(chan concurrentStream,3)
	byteArrayArrays := [3][]byte{}
	go createArray(channel,stream.S1,1)
	go createArray(channel,stream.S2,2)
	go createArray(channel,stream.S3,3)

	for i:=0;i<3;i++{
		result := <-channel
		byteArrayArrays[result.id-1]=result.stream
	}
	result := createOutput(stream.BV,byteArrayArrays)
	return result[0:len(result)-int(stream.emptyBytes)]
}