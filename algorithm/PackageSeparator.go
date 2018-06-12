package algorithm

import (
	"errors"
	"math"
)

var ErrOutOfRange = errors.New("bytes count is out of range")

func SeparatePackage(packBytes []byte, owner uint32, start uint32, key uint32, packDataLength uint16) ([]DataPackage, uint32, error) {
	packBytesLength := len(packBytes)
	if packBytesLength > math.MaxUint16*math.MaxUint16 {
		return nil, start, ErrOutOfRange
	}
	packCount := packBytesLength / int(packDataLength)
	if packBytesLength%int(packDataLength) > 0 {
		packCount++
	}
	packList := make([]DataPackage, packCount)
	for index := 0; index < packCount; index++ {
		dataStart := index * int(packDataLength)
		dataEnd := dataStart + int(packDataLength)
		if dataEnd >= packBytesLength {
			dataEnd = packBytesLength
		}
		pack := &packList[index]

		pack.Owner = owner
		pack.Index = start
		pack.Key = key
		pack.Length = uint16(dataEnd - dataStart)
		pack.KIndex = uint16(index)
		pack.KCount = uint16(packCount)
		pack.Data = packBytes[dataStart:dataEnd]
		start++
	}
	return packList, start, nil
}
