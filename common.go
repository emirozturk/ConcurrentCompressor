package main
/*

import (
	"encoding/binary"
	"strconv"
)

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
 */