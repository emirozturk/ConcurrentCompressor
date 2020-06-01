package main

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"sync"
)

type cc struct {
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

var frequencyMap = make(map[string]*int)
var sortedDictionary []kv

var lock = sync.RWMutex{}

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

func extractNgrams(wg *sync.WaitGroup, block []byte, ngramSize int) {
	defer wg.Done()
	length := len(block)
	for i := 0; i < length; i += ngramSize {
		key := string(block[i*ngramSize : (i+1)*ngramSize])
		lock.Lock()
		if frequencyMap[key] == nil {
			frequencyMap[key] = new(int)
		}
		*frequencyMap[key]++
		lock.Unlock()
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
	for i := first; i < last; i++ {
		dictionary.dictionary[sortedDictionary[i].Key] = i + 1
	}
	dictionary.id = index
	channel <- dictionary
}
func createStream(channel chan concurrentStream, inputBytes []byte, ngramSize int, D1 map[string]int, D2 map[string]int, index int) {
	var result concurrentStream
	if index == 1 {
		for i := 0; i < len(inputBytes); i += ngramSize {
			key := string(inputBytes[i*ngramSize : (i+1)*ngramSize])
			val, ok := D1[key]
			if ok && val != 0 {
				result.stream = append(result.stream, byte(val))
			}
		}
	} else if index == 2 {
		for i := 0; i < len(inputBytes); i += ngramSize {
			key := string(inputBytes[i*ngramSize : (i+1)*ngramSize])
			val, ok := D2[key]
			if ok {
				output := make([]byte, 2)
				binary.BigEndian.PutUint16(output, uint16(val))
				result.stream = append(result.stream, output...)
			}
		}
	} else if index == 3 {
		for i := 0; i < len(inputBytes); i += ngramSize {
			bytes := inputBytes[i*ngramSize : (i+1)*ngramSize]
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

	for i := 0; i < len(inputBytes); i += ngramSize {
		key := string(inputBytes[i*ngramSize : (i+1)*ngramSize])
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
	result.id = index
	channel <- result
}
func createDictionaryStream(channel chan concurrentStream, dictionary map[string]int, index int) {
	var result concurrentStream

	////////

	result.id = index
	channel <- result
}
func compress(fileName string, ngramSize int) {
	cpuCount := runtime.NumCPU()
	dataStream, _ := os.Open(fileName)
	inputBytes, _ := ioutil.ReadAll(dataStream)
	blockSize := len(inputBytes) / cpuCount

	var wg sync.WaitGroup

	for i := 0; i < cpuCount; i++ {
		go extractNgrams(&wg, inputBytes[i*blockSize:(i+1)*blockSize], ngramSize)
		wg.Add(1)
	}

	wg.Wait()

	createSortedDictionary()

	dictionaries := make(chan concurrentDictionary, 2)

	go createDictionary(dictionaries, 0, 255, 0)
	go createDictionary(dictionaries, 255, 255+65536, 1)

	dictionaryArray := make([]map[string]int, 2)
	rd1 := <-dictionaries
	rd2 := <-dictionaries
	dictionaryArray[rd1.id] = rd1.dictionary
	dictionaryArray[rd2.id] = rd2.dictionary
	D1 := dictionaryArray[0]
	D2 := dictionaryArray[1]

	streams := make(chan concurrentStream, 4)

	go createStream(streams, inputBytes, ngramSize, D1, D2, 1)
	go createStream(streams, inputBytes, ngramSize, D1, D2, 2)
	go createStream(streams, inputBytes, ngramSize, D1, D2, 3)
	go createBV(streams, inputBytes, ngramSize, D1, D2, 4)
	go createDictionaryStream(streams, D1, 5)
	go createDictionaryStream(streams, D2, 6)

	streamArray := make([][]byte, 7)
	sd1 := <-streams
	sd2 := <-streams
	sd3 := <-streams
	sd4 := <-streams
	sd5 := <-streams
	sd6 := <-streams
	streamArray[sd1.id] = sd1.stream
	streamArray[sd2.id] = sd2.stream
	streamArray[sd3.id] = sd3.stream
	streamArray[sd4.id] = sd4.stream
	streamArray[sd5.id] = sd5.stream
	streamArray[sd6.id] = sd6.stream

	var outputStream cc
	outputStream.S1 = streamArray[1]
	outputStream.S2 = streamArray[2]
	outputStream.S3 = streamArray[3]
	outputStream.BV = streamArray[4]
	outputStream.D1 = streamArray[5]
	outputStream.D2 = streamArray[6]
}
