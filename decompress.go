package main

import (
	"encoding/binary"
	"sync"
)

var D1 [256][]byte
var D2 [65536][]byte
var ngramSize int

func createOutput(bvBytes []byte,streams [3][][]byte) []byte{
	var output []byte
	bv := bytesToBools(bvBytes[1:])
	reduntantBits:=bvBytes[0]
	length:=len(bv)-int(reduntantBits)
	var s1Counter,s2Counter,s3Counter int
	for i:=0;i<length;i++{
		if !bv[i]{
			output = append(output, streams[0][s1Counter]...)
			s1Counter++
		} else if bv[i] {
			if !bv[i+1] {
				output = append(output, streams[1][s2Counter]...)
				s2Counter++
			}else{
				output = append(output, streams[2][s3Counter]...)
				s3Counter++
			}
			i++
		}
	}
	return output
}

func createDictionaryArray(wg *sync.WaitGroup,stream []byte,index int){
	defer wg.Done()
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
	var output []byte

	var wg sync.WaitGroup

	ngramSize =int(stream.D1[0])
	go createDictionaryArray(&wg,stream.D1,1)
	wg.Add(1)
	go createDictionaryArray(&wg,stream.D2,2)
	wg.Add(1)

	wg.Wait()

	channel := make(chan concurrentByteArray,3)
	byteArrayArrays := [3][][]byte{}
	go createArray(channel,stream.S1,1)
	go createArray(channel,stream.S2,2)
	go createArray(channel,stream.S3,3)

	for i:=0;i<3;i++{
		result := <-channel
		byteArrayArrays[result.id-1]=result.array
	}

	output = createOutput(stream.BV,byteArrayArrays)

	return output
}