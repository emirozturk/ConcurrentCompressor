package main

import "sync"

var D1 [256]string
var D2 [65536]string
var ngramSize int

func createOutput(bvBytes []byte,streams [3][]string) []byte{
	var output []byte
	bv := bytesToBools(bvBytes)
	var s1Counter,s2Counter,s3Counter int
	for i:=0;i<len(bv);i++{
		if !bv[i]{
				output = append(output, []byte(streams[0][s1Counter])...)
				s1Counter++
		} else if bv[i] {
			if !bv[i+1] {
				output = append(output, []byte(streams[1][s2Counter])...)
				s2Counter++
			}else{
				output = append(output, []byte(streams[2][s3Counter])...)
				s3Counter++
			}
		}
	}
	return output
}

func createDictionaryArray(wg *sync.WaitGroup,stream []byte,index int){
	defer wg.Done()
	counter := 0
	if index == 1 {
		for i:=1;i<len(stream);i+=ngramSize{
			D1[counter] = string(stream[i:i+ngramSize])
			counter++
		}
	} else if index == 2 {
		for i:=0;i<len(stream);i+=ngramSize{
			D2[counter] = string(stream[i:i+ngramSize])
			counter++
		}
	}
}
func createArray(channel chan concurrentString,stream []byte,index int){
	cc := concurrentString{}
	cc.id = index

	for i:=0;i<len(stream);i+=ngramSize{
		cc.array = append(cc.array, string(stream[i:i+ngramSize]))
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

	channel := make(chan concurrentString,3)
	stringArrays := [3][]string{}
	go createArray(channel,stream.S1,1)
	go createArray(channel,stream.S2,2)
	go createArray(channel,stream.S3,3)

	for i:=0;i<3;i++{
		result := <-channel
		stringArrays[result.id]=result.array
	}

	output = createOutput(stream.BV,stringArrays)

	return output
}
