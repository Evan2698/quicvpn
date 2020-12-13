package utils

import (
	"bytes"
	"encoding/binary"
)

//Int2Bytes ...
func Int2Bytes(n uint32) []byte {
	u := uint32(n)
	var hello bytes.Buffer
	binary.Write(&hello, binary.BigEndian, u)
	return hello.Bytes()
}

//Bytes2Int ...
func Bytes2Int(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp
}
