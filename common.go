package main

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
	Key   uint64
	Value int
}
type concurrentStream struct {
	id     int
	stream []byte
}
type concurrentDictionary struct {
	id         int
	dictionary map[uint64]int
}
func uint32ToByteArray(intValue uint32) []byte {
	buffer := make([]byte, strconv.IntSize/8)
	binary.BigEndian.PutUint32(buffer, intValue)
	return buffer
}
func byteArrayToUint32(byteArray []byte) uint32 {
	return binary.BigEndian.Uint32(byteArray)
}
func uint64ToByteArray(intValue uint64) []byte {
	buffer := make([]byte, strconv.IntSize*2/8)
	binary.BigEndian.PutUint64(buffer, intValue)
	return buffer
}
func byteArrayToUint64(byteArray []byte) uint64 {
	array := make([]byte,8)
	copy(array,byteArray)
	return binary.BigEndian.Uint64(array)
}