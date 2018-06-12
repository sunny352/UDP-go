package Pool_test

import (
	"log"
	"reflect"
	"testing"
	"unsafe"
)

func TestMem(t *testing.T) {
	temp := make([]byte, 1024)
	temp1 := temp[:512]
	temp2 := temp[512:]

	tempHead := (*reflect.SliceHeader)(unsafe.Pointer(&temp))
	log.Println(tempHead.Data)

	log.Println((*reflect.SliceHeader)(unsafe.Pointer(&temp)))
	log.Println((*reflect.SliceHeader)(unsafe.Pointer(&temp1)))
	log.Println((*reflect.SliceHeader)(unsafe.Pointer(&temp2)))
}
