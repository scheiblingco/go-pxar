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

func GetSymlinkTest() pxar.PxarSymlink {
	return pxar.PxarSymlink{
		Target: "file.txt",
	}
}

func TestWriteSymlink(t *testing.T) {
	pf := GetSymlinkTest()

	wantLength := 25
	wantData := bytes.NewBuffer([]byte{})

	// PXAR_ENTRY Descriptor
	AppendInterface(pxar.PXAR_SYMLINK, wantData, t)

	// Length of record
	AppendInterface(uint64(25), wantData, t)

	// Filename
	AppendInterface([]byte("file.txt"), wantData, t)

	// Terminator
	AppendInterface(byte(0x0), wantData, t)

	wantBytes := wantData.Bytes()

	actual := bytes.NewBuffer([]byte{})

	pos := uint64(0)

	num, err := pf.Write(actual, &pos)
	if err != nil {
		t.Errorf("an error occured while writing the entry %v: %e", pf, err)
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

func TestStreamSymlink(t *testing.T) {
	pe := GetSymlinkTest()

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

func TestSymlinkChan(t *testing.T) {
	pe := GetSymlinkTest()

	buf := bytes.NewBuffer([]byte{})
	bufpos := uint64(0)

	ch := make(chan []byte)
	done := make(chan error)
	chanpos := uint64(0)
	chanres := []byte{}

	pe.Write(buf, &bufpos)

	go func() {
	L:
		for {
			select {
			case res := <-ch:
				chanres = append(chanres, res...)
			case don := <-done:
				if don != nil {
					t.Errorf("an error occured while writing the symlink %v: %e", pe, don)
				}
				break L
			default:
				continue
			}
		}
	}()

	pe.WriteChannel(ch, &chanpos)
	done <- nil

	bufBytes := buf.Bytes()

	for i := range bufBytes {
		if bufBytes[i] != chanres[i] {
			t.Errorf("mismatch at position %d, expected %b but got %b", i, bufBytes[i], chanres[i])
		}
	}

	close(ch)
}
