package main

import (
	"encoding/binary"
	"math"
	"runtime"
	"sort"
)

var frequencyMap = make(map[string]*int)
var sortedDictionary []kv

func extractNgrams(channel chan map[string]*int, block []byte, ngramSize int) {
	fMap := make(map[string]*int)
	length := len(block)
	for i := 0; i < length; i += ngramSize {
		ngram := block[i : i+ngramSize]
		key := string(ngram)
		if fMap[key] == nil {
			fMap[key] = new(int)
		}
		*fMap[key]++
	}
	channel <- fMap
}
func createFrequencyMap(array []map[string]*int) {
	for _, ngrams := range array {
		for key, value := range ngrams {
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
			key := string(inputBytes[i : i+ngramSize])
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
			bytes := inputBytes[i : i+ngramSize]
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
	for i := 0; i < length; i += ngramSize {
		key := string(inputBytes[i : i+ngramSize])
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
	redunantBits := byte(len(result.stream)*8 - len(vector))
	result.stream = append([]byte{redunantBits}, result.stream...)
	result.id = index
	channel <- result
}
func createDictionaryStream(channel chan concurrentStream, ngramSize int, index int) {
	var result concurrentStream
	min := 0
	max := 0
	if index == 5 {
		min = 0
		max = int(math.Min(256, float64(len(sortedDictionary))))
		result.stream = append(result.stream, byte(ngramSize))
	} else if index == 6 {
		min = 256
		max = int(math.Min(65536+256, float64(len(sortedDictionary))))
	}
	for i := min; i < max; i++ {
		result.stream = append(result.stream, []byte(sortedDictionary[i].Key)...)
	}
	result.id = index
	channel <- result
}
func compress(inputBytes []byte, ngramSize int) ccStream {
	cpuCount := runtime.NumCPU()
	blockSize := len(inputBytes) / cpuCount

	frequencies := make(chan map[string]*int, cpuCount)
	for i := 0; i < cpuCount; i++ {
		go extractNgrams(frequencies, inputBytes[i*blockSize:(i+1)*blockSize], ngramSize)
	}

	frequencyArray := make([]map[string]*int, cpuCount)
	for i := 0; i < cpuCount; i++ {
		frequencyArray[i] = <-frequencies
	}
	createFrequencyMap(frequencyArray)

	createSortedDictionary()

	dictionaries := make(chan concurrentDictionary, 2)

	go createDictionary(dictionaries, 0, 256, 0)
	go createDictionary(dictionaries, 256, 256+65536, 1)

	dictionaryArray := [2]map[string]int{}
	for i := 0; i < 2; i++ {
		buffer := <-dictionaries
		dictionaryArray[buffer.id] = buffer.dictionary
	}
	D1 := dictionaryArray[0]
	D2 := dictionaryArray[1]

	streams := make(chan concurrentStream, 4)

	for i := 1; i < 4; i++ {
		go createStream(streams, inputBytes, ngramSize, D1, D2, i)
	}

	go createBV(streams, inputBytes, ngramSize, D1, D2, 4)

	for i := 5; i < 7; i++ {
		go createDictionaryStream(streams, ngramSize, i)
	}

	streamArray := [7][]byte{}
	for i := 1; i < 7; i++ {
		cs := <-streams
		streamArray[cs.id] = cs.stream
	}
	outputStream := ccStream{}
	outputStream.S1 = streamArray[1]
	outputStream.S2 = streamArray[2]
	outputStream.S3 = streamArray[3]
	outputStream.BV = streamArray[4]
	outputStream.D1 = streamArray[5]
	outputStream.D2 = streamArray[6]
	return outputStream
}
