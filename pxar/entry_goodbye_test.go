package pxar_test

import (
	"bytes"
	"testing"

	"github.com/scheiblingco/go-pxar/pxar"
)

func GetGoodbyeTest() pxar.PxarGoodbye {
	return pxar.PxarGoodbye{
		FolderStart:  0x0,
		GoodbyeStart: 40,
		Items: []pxar.GoodbyeItem{
			{
				Hash:   0x300,
				Offset: 10,
				Length: 10,
			},
			{
				Hash:   0x200,
				Offset: 20,
				Length: 10,
			},
			{
				Hash:   0x100,
				Offset: 30,
				Length: 10,
			},
		},
	}
}

func TestWriteGoodbye(t *testing.T) {
	pf := GetGoodbyeTest()

	wantData := bytes.NewBuffer([]byte{})

	AppendInterface(pxar.PXAR_GOODBYE, wantData, t)
	AppendInterface(uint64(112), wantData, t)

	AppendInterface(pxar.GoodbyeItem{
		Hash:   0x200,
		Offset: 30,
		Length: 10,
	}, wantData, t)

	AppendInterface(pxar.GoodbyeItem{
		Hash:   0x100,
		Offset: 20,
		Length: 10,
	}, wantData, t)

	AppendInterface(pxar.GoodbyeItem{
		Hash:   0x300,
		Offset: 40,
		Length: 10,
	}, wantData, t)

	AppendInterface(pxar.GoodbyeItem{
		Hash:   pxar.PXAR_GOODBYE_TAIL_MARKER,
		Offset: 40,
		Length: 112,
	}, wantData, t)

	wantLength := 112
	wantBytes := wantData.Bytes()

	actual := bytes.NewBuffer([]byte{})
	pos := uint64(0)

	num, err := pf.Write(actual, &pos)
	if err != nil {
		t.Errorf("an error occured while writing the entry %v: %e", pf, err)
	}

	if num != uint64(wantLength) {
		t.Errorf("length mismatch, expected %d but got %d", 40, num)
	}

	actualBytes := actual.Bytes()

	for i := range wantBytes {
		if actualBytes[i] != wantBytes[i] {
			t.Errorf("mismatch at position %d, expected %b but got %b", i, wantBytes[i], actualBytes[i])
		}
	}
}

func TestStreamGoodbye(t *testing.T) {
	pe := GetGoodbyeTest()

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

func TestGoodbyeChan(t *testing.T) {
	pe := GetGoodbyeTest()

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

// func TestSymlinkChan(t *testing.T) {
// 	pe := GetSymlinkTest()

// 	buf := bytes.NewBuffer([]byte{})
// 	bufpos := uint64(0)

// 	ch := make(chan []byte)
// 	done := make(chan error)
// 	chanpos := uint64(0)
// 	chanres := []byte{}

// 	pe.Write(buf, &bufpos)

// 	go func() {
// 	L:
// 		for {
// 			select {
// 			case res := <-ch:
// 				chanres = append(chanres, res...)
// 			case don := <-done:
// 				if don != nil {
// 					t.Errorf("an error occured while writing the symlink %v: %e", pe, don)
// 				}
// 				break L
// 			default:
// 				continue
// 			}
// 		}
// 	}()

// 	pe.WriteChannel(ch, &chanpos)
// 	done <- nil

// 	bufBytes := buf.Bytes()

// 	for i := range bufBytes {
// 		if bufBytes[i] != chanres[i] {
// 			t.Errorf("mismatch at position %d, expected %b but got %b", i, bufBytes[i], chanres[i])
// 		}
// 	}

// 	close(ch)
// }
