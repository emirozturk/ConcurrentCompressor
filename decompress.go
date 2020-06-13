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
	reduntantBits:=bvBytes[0]
	length:=bitCount-int(reduntantBits)

	mask := [4]byte{192,48,12,3}
	maskResults :=[4][3]byte{{0,128,192},{0,32,48},{0,8,12},{0,2,3}}
	var s1Counter,s2Counter,s3Counter,outputCounter,shiftCounter,bvCounter int
	bvCounter=1
	for i:=0;i<length;i+=2{
		bitResult := bvBytes[bvCounter] & mask[shiftCounter]
		if  bitResult == maskResults[shiftCounter][0]  {
			copy(output[outputCounter:],streams[0][s1Counter*ngramSize:(s1Counter+1)*ngramSize])
			s1Counter++
		} else if bitResult == maskResults[shiftCounter][1] {
			copy(output[outputCounter:],streams[1][s2Counter*ngramSize:(s2Counter+1)*ngramSize])
			s2Counter++
		} else{
			copy(output[outputCounter:],streams[2][s3Counter*ngramSize:(s3Counter+1)*ngramSize])
			s3Counter++
		}
		outputCounter+=ngramSize
		shiftCounter++
		if shiftCounter==4 {
			bvCounter++
			shiftCounter=0
		}
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

	return createOutput(stream.BV,byteArrayArrays)
}