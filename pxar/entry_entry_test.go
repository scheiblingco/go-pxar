package pxar_test

import (
	"bytes"
	"testing"

	"github.com/scheiblingco/go-pxar/pxar"
)

// type PxarEntry struct {
// 	Mode         uint64
// 	Flags        uint64
// 	Uid          uint32
// 	Gid          uint32
// 	MtimeSecs    uint64
// 	MtimeNanos   uint32
// 	MtimePadding uint32
// }

func GetEntryTest() pxar.PxarEntry {
	return pxar.PxarEntry{
		Mode:         0o40755,
		Flags:        0x0,
		Uid:          1000,
		Gid:          1000,
		MtimeSecs:    12345678,
		MtimeNanos:   987654321,
		MtimePadding: 0x0,
	}
}

func TestEntryWrite(t *testing.T) {
	pe := GetEntryTest()

	wantLength := 56
	wantData := bytes.NewBuffer([]byte{})

	// PXAR_ENTRY Descriptor
	AppendInterface(pxar.PXAR_ENTRY, wantData, t)

	// Length of record
	AppendInterface(uint64(56), wantData, t)

	// Mode
	AppendInterface(uint64(0o40755), wantData, t)

	// Flags
	AppendInterface(uint64(0x0), wantData, t)

	// Uid
	AppendInterface(uint32(1000), wantData, t)

	// Gid
	AppendInterface(uint32(1000), wantData, t)

	// MtimeSecs
	AppendInterface(uint64(12345678), wantData, t)

	// MtimeNanos
	AppendInterface(uint32(987654321), wantData, t)

	// MtimePadding
	AppendInterface(uint32(0x0), wantData, t)

	wantBytes := wantData.Bytes()

	actual := bytes.NewBuffer([]byte{})

	pos := uint64(0)

	num, err := pe.Write(actual, &pos)
	if err != nil {
		t.Errorf("an error occured while writing the entry %v: %e", pe, err)
	}

	if num != uint64(wantLength) {
		t.Errorf("length mismatch, expected %d but got %d", wantLength, num)
	}

	actualBytes := actual.Bytes()

	for i := range wantBytes {
		if actualBytes[i] != wantBytes[i] {
			t.Errorf("mismatch at position %d, expected %b but got %b", i, wantBytes[i], actualBytes[i])
		}
	}
}

func TestEntryStream(t *testing.T) {
	pe := GetEntryTest()

	buf := bytes.NewBuffer([]byte{})
	bufpos := uint64(0)

	streambuf := bytes.NewBuffer([]byte{})
	streambufpos := uint64(0)

	pe.Write(buf, &bufpos)
	pe.WriteStream(streambuf, &streambufpos)

	if bufpos != streambufpos {
		t.Errorf("mismatch between buf and stream writers: %d on buf vs %d on stream", bufpos, streambufpos)
	}

	bufBytes := buf.Bytes()
	streamBytes := streambuf.Bytes()

	for i := range bufBytes {
		if bufBytes[i] != streamBytes[i] {
			t.Errorf("mismatch at position %d, expected %b but got %b", i, bufBytes[i], streamBytes[i])
		}
	}
}
