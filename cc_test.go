package main

import (
	"encoding/binary"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)


var frequencyMap = make(map[string]*int)
var sortedDictionary []kv

func extractNgrams(channel chan map[string]*int, block []byte, ngramSize int) {
	fMap := make(map[string]*int)
	length := len(block)
	for i := 0; i < length; i += ngramSize {
		ngram := block[i: i+ngramSize]
		key := string(ngram)
		if fMap[key] == nil {
			fMap[key] = new(int)
		}
		*fMap[key]++
	}
	channel<-fMap
}
func createFrequencyMap(array []map[string]*int) {
	for _,ngrams := range array{
		for key,value := range ngrams{
			if frequencyMap[key] == nil {
				frequencyMap[key] = new(int)
			}
			*frequencyMap[key] += *value
		}
	}
}
func createSortedDictionary() {
	for k, v := range frequencyMap {
		sortedDictionary = append(sortedDictionary, kv{k, v})
	}
	sort.Slice(sortedDictionary, func(i, j int) bool {
		return *sortedDictionary[i].Value > *sortedDictionary[j].Value
	})
}
func createDictionary(channel chan concurrentDictionary, first int, last int, index int) {
	var dictionary concurrentDictionary
	dictionary.dictionary = make(map[string]int)
	last = int(math.Min(float64(len(sortedDictionary)), float64(last)))
	for i := first; i < last; i++ {
		dictionary.dictionary[sortedDictionary[i].Key] = i
	}
	dictionary.id = index
	channel <- dictionary
}
func createStream(channel chan concurrentStream, inputBytes []byte, ngramSize int, D1 map[string]int, D2 map[string]int, index int) {
	var result concurrentStream
	length := len(inputBytes)
	if index == 1 {
		for i := 0; i < length; i += ngramSize {
			key := string(inputBytes[i:i+ngramSize])
			val, ok := D1[key]
			if ok {
				result.stream = append(result.stream, byte(val))
			}
		}
	} else if index == 2 {
		for i := 0; i < length; i += ngramSize {
			key := string(inputBytes[i : i+ngramSize])
			val, ok := D2[key]
			if ok {
				output := make([]byte, 2)
				binary.BigEndian.PutUint16(output, uint16(val-256))
				result.stream = append(result.stream, output...)
			}
		}
	} else if index == 3 {
		for i := 0; i < length; i += ngramSize {
			bytes := inputBytes[i :i+ngramSize]
			key := string(bytes)
			_, okD1 := D1[key]
			_, okD2 := D2[key]
			if !okD1 && !okD2 {
				result.stream = append(result.stream, bytes...)
			}
		}
	}
	result.id = index
	channel <- result
}
func createBV(channel chan concurrentStream, inputBytes []byte, ngramSize int, D1 map[string]int, D2 map[string]int, index int) {
	var result concurrentStream
	var vector []bool
	length := len(inputBytes)
	for i := 0; i <length;  i += ngramSize {
		key := string(inputBytes[i:i+ngramSize])
		_, ok := D1[key]
		if ok {
			vector = append(vector, false)
		} else {
			_, ok := D2[key]
			if ok {
				vector = append(vector, true, false)
			} else {
				vector = append(vector, true, true)
			}
		}
	}
	result.stream = boolsToBytes(vector)
	redunantBits:=byte(len(result.stream)*8 - len(vector))
	result.stream = append([]byte{redunantBits},result.stream...)
	result.id = index
	channel <- result
}
func createDictionaryStream(channel chan concurrentStream,ngramSize int, index int) {
	var result concurrentStream
	min:=0
	max:=0
	if index == 5{
		min = 0
		max = int(math.Min(256, float64(len(sortedDictionary))))
		result.stream = append(result.stream,byte(ngramSize))
	}else if index == 6{
		min = 256
		max = int(math.Min(65536+256, float64(len(sortedDictionary))))
	}
	for i:=min;i<max;i++{
		result.stream = append(result.stream,[]byte(sortedDictionary[i].Key)...)
	}
	result.id = index
	channel <- result
}
func compress(inputBytes []byte, ngramSize int) ccStream {
	cpuCount := runtime.NumCPU()
	blockSize := len(inputBytes) / cpuCount

	frequencies := make(chan map[string]*int,cpuCount)
	for i := 0; i < cpuCount; i++ {
		go extractNgrams(frequencies, inputBytes[i*blockSize:(i+1)*blockSize], ngramSize)
	}

	frequencyArray := make([]map[string]*int,cpuCount)
	for i:=0;i<cpuCount;i++{
		frequencyArray[i] = <-frequencies
	}
	createFrequencyMap(frequencyArray)

	createSortedDictionary()

	dictionaries := make(chan concurrentDictionary, 2)

	go createDictionary(dictionaries, 0, 256, 0)
	go createDictionary(dictionaries, 256, 256+65536, 1)

	dictionaryArray := [2]map[string]int{}
	for i:=0;i<2;i++{
		buffer:=<-dictionaries
		dictionaryArray[buffer.id] = buffer.dictionary
	}
	D1 := dictionaryArray[0]
	D2 := dictionaryArray[1]

	streams := make(chan concurrentStream, 4)

	for i:=1;i<4;i++{
		go createStream(streams, inputBytes, ngramSize, D1, D2, i)
	}

	go createBV(streams, inputBytes, ngramSize, D1, D2, 4)

	for i:=5;i<7;i++{
		go createDictionaryStream(streams,ngramSize, i)
	}

	streamArray := [7][]byte{}
	for i:=1;i<7;i++{
		cs := <-streams
		streamArray[cs.id] = cs.stream
	}
	outputStream:= ccStream{}
	outputStream.S1 = streamArray[1]
	outputStream.S2 = streamArray[2]
	outputStream.S3 = streamArray[3]
	outputStream.BV = streamArray[4]
	outputStream.D1 = streamArray[5]
	outputStream.D2 = streamArray[6]
	return outputStream
}

var D1 [256]string
var D2 [65536]string
var ngramSize int

func createOutput(bvBytes []byte,streams [3][]string) []byte{
	var output []byte
	bv := bytesToBools(bvBytes[1:])
	reduntantBits:=bvBytes[0]
	length:=len(bv)-int(reduntantBits)
	var s1Counter,s2Counter,s3Counter int
	for i:=0;i<length;i++{
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
	if index == 1{
		cc.array = make([]string,0,len(stream))
		for i:=0;i<len(stream);i++{
			cc.array = append(cc.array, D1[stream[i]])
		}
	}else if index ==2{
		cc.array = make([]string,0,len(stream)/2)
		for i:=0;i<len(stream);i+=2{
			cc.array = append(cc.array, D2[binary.BigEndian.Uint16(stream[i:i+2])])
		}
	}else if index == 3{
		cc.array = make([]string,0,len(stream)/ngramSize)
		for i:=0;i<len(stream);i+=ngramSize{
			cc.array = append(cc.array, string(stream[i:i+ngramSize]))
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

	channel := make(chan concurrentString,3)
	stringArrays := [3][]string{}
	go createArray(channel,stream.S1,1)
	go createArray(channel,stream.S2,2)
	go createArray(channel,stream.S3,3)

	for i:=0;i<3;i++{
		result := <-channel
		stringArrays[result.id-1]=result.array
	}

	output = createOutput(stream.BV,stringArrays)

	return output
}

func readFile(fileName string) []byte {
	dataStream, _ := os.Open(fileName)
	inputBytes, _ := ioutil.ReadAll(dataStream)
	_ = dataStream.Close()
	return inputBytes
}
func writeFile(fileName string,stream []byte){
	dataStream,_ := os.Create(strings.Replace(fileName,".ccf",".cca",1))
	_, _ = dataStream.Write(stream)
	_ = dataStream.Close()
}
func readCCStream(fileName string) ccStream{
	size:= strconv.IntSize/8
	stream :=ccStream{}
	dataStream ,_ := os.Open(fileName)
	input, _ := ioutil.ReadAll(dataStream)
	lenArray := [6]int{}
	streamArray := [6][]byte{}
	for i:=0;i<6;i++{
		lenArray[i] = int(byteArrayToUint32(input[i*size : (i+1)*size]))
	}
	offset := 6*size
	for i:=0;i<6;i++{
		streamArray[i] = input[offset:offset+lenArray[i]]
		offset+=lenArray[i]
	}
	stream.D1 = streamArray[0]
	stream.D2 = streamArray[1]
	stream.S1 = streamArray[2]
	stream.S2 = streamArray[3]
	stream.S3 = streamArray[4]
	stream.BV = streamArray[5]
	return stream
}
func writeCCStream(inputName string,stream ccStream){
	var output []byte
	output = append(output, uint32ToByteArray(uint32(len(stream.D1)))...)
	output = append(output, uint32ToByteArray(uint32(len(stream.D2)))...)
	output = append(output, uint32ToByteArray(uint32(len(stream.S1)))...)
	output = append(output, uint32ToByteArray(uint32(len(stream.S2)))...)
	output = append(output, uint32ToByteArray(uint32(len(stream.S3)))...)
	output = append(output, uint32ToByteArray(uint32(len(stream.BV)))...)

	output = append(output, stream.D1...)
	output = append(output, stream.D2...)
	output = append(output, stream.S1...)
	output = append(output, stream.S2...)
	output = append(output, stream.S3...)
	output = append(output, stream.BV...)

	dataStream, _ := os.Create(inputName+".ccf")
	_, _ = dataStream.Write(output)
	_ = dataStream.Close()
}

type ccStream struct {
	D1 []byte
	D2 []byte
	S1 []byte
	S2 []byte
	S3 []byte
	BV []byte
}
type kv struct {
	Key   string
	Value *int
}
type concurrentStream struct {
	id     int
	stream []byte
}
type concurrentDictionary struct {
	id         int
	dictionary map[string]int
}
type concurrentString struct{
	id int
	array []string
}
func uint32ToByteArray(intValue uint32) []byte {
	buffer := make([]byte, strconv.IntSize/8)
	binary.BigEndian.PutUint32(buffer, intValue)
	return buffer
}
func byteArrayToUint32(byteArray []byte) uint32 {
	return binary.BigEndian.Uint32(byteArray)
}

func boolsToBytes(t []bool) []byte {
	b := make([]byte, (len(t)+7)/8)
	for i, x := range t {
		if x {
			b[i/8] |= 0x80 >> uint(i%8)
		}
	}
	return b
}

func bytesToBools(b []byte) []bool {
	t := make([]bool, 8*len(b))
	for i, x := range b {
		for j := 0; j < 8; j++ {
			if (x<<uint(j))&0x80 == 0x80 {
				t[8*i+j] = true
			}
		}
	}
	return t
}

func TestCompress(t *testing.T) {
	fileName := "C:/Users/emiro/Desktop/english.50MB"
	ngramSize := 4
	stream := compress(readFile(fileName), ngramSize)
	writeCCStream(fileName,stream)
}
func TestDecompress(t *testing.T) {
	fileName := "C:/Users/emiro/Desktop/english.50MB.ccf"
	output := decompress(readCCStream(fileName))
	writeFile(fileName,output)
}
