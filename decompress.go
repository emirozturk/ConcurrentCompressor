package main

import (
	"encoding/binary"
)

var D1 [256][]byte
var D2 [65536][]byte
var ngramSize int

func createOutput(bvBytes []byte,streams [3][][]byte) []byte{
	bv := bytesToBools(bvBytes[1:])
	output := make([]byte,len(bv)*4/2)
	reduntantBits:=bvBytes[0]
	length:=len(bv)-int(reduntantBits)
	var s1Counter,s2Counter,s3Counter,outputCounter int
	for i:=0;i<length;i+=2{
		if !bv[i]{
			copy(output[outputCounter:],streams[0][s1Counter])
			s1Counter++
		} else if bv[i] {
			if !bv[i+1] {
				copy(output[outputCounter:],streams[1][s2Counter])
				s2Counter++
			}else{
				copy(output[outputCounter:],streams[2][s3Counter])
				s3Counter++
			}
		}
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
func createArray(channel chan concurrentByteArray,stream []byte,index int){
	cc := concurrentByteArray{}
	cc.id = index
	if index == 1{
		cc.array = make([][]byte,0,len(stream))
		for i:=0;i<len(stream);i++{
			cc.array = append(cc.array, D1[stream[i]])
		}
	}else if index ==2{
		cc.array = make([][]byte,0,len(stream)/2)
		for i:=0;i<len(stream);i+=2{
			cc.array = append(cc.array, D2[binary.BigEndian.Uint16(stream[i:i+2])])
		}
	}else if index == 3{
		cc.array = make([][]byte,0,len(stream)/ngramSize)
		for i:=0;i<len(stream);i+=ngramSize{
			cc.array = append(cc.array, stream[i:i+ngramSize])
		}
	}

	channel <-cc
}
func decompress(stream ccStream) []byte {
	ngramSize =int(stream.D1[0])
	createDictionaryArray(stream.D1,1)
	createDictionaryArray(stream.D2,2)

	channel := make(chan concurrentByteArray,3)
	byteArrayArrays := [3][][]byte{}
	go createArray(channel,stream.S1,1)
	go createArray(channel,stream.S2,2)
	go createArray(channel,stream.S3,3)

	for i:=0;i<3;i++{
		result := <-channel
		byteArrayArrays[result.id-1]=result.array
	}

	return createOutput(stream.BV,byteArrayArrays)
}