package pxar_test

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/scheiblingco/go-pxar/pxar"
)

func AppendInterface(ifa interface{}, buf *bytes.Buffer, t *testing.T) {
	err := binary.Write(buf, binary.LittleEndian, ifa)
	if err != nil {
		t.Errorf("Error writing %v: %e", ifa, err)
	}
}

func TestInterfaces(t *testing.T) {
	ifa := reflect.TypeOf((*pxar.PxarSection)(nil)).Elem()

	// PxarEntry
	if !reflect.TypeOf(&pxar.PxarEntry{}).Implements(ifa) {
		t.Errorf("PxarEntry does not implement PxarSection")
	}

	// PxarFilename
	if !reflect.TypeOf(&pxar.PxarFilename{}).Implements(ifa) {
		t.Errorf("PxarFilename does not implement PxarSection")
	}

	// PxarPayload
	if !reflect.TypeOf(&pxar.PxarPayload{}).Implements(ifa) {
		t.Errorf("PxarPayload does not implement PxarSection")
	}

	// PxarSymlink
	if !reflect.TypeOf(&pxar.PxarSymlink{}).Implements(ifa) {
		t.Errorf("PxarSymlink does not implement PxarSection")
	}

	// PxarGoodbye
	if !reflect.TypeOf(&pxar.PxarGoodbye{}).Implements(ifa) {
		t.Errorf("PxarGoodbye does not implement PxarSection")
	}
}
