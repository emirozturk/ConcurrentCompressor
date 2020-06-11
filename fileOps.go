package main
/*

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

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
*/