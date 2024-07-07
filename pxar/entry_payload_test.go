package pxar_test

import (
	"bytes"
	"testing"

	"github.com/scheiblingco/go-pxar/pxar"
)

func GetPayloadTest() pxar.PxarPayload {
	strm := bytes.NewBuffer([]byte("qwertyuiopåäölkjhgfdsapq"))

	return pxar.PxarPayload{
		Size:   uint64(strm.Len()),
		Stream: strm,
	}
}

func TestWritePayload(t *testing.T) {
	pp := GetPayloadTest()

	wantLength := 43
	wantData := bytes.NewBuffer([]byte{})

	// PXAR_ENTRY Descriptor
	AppendInterface(pxar.PXAR_PAYLOAD, wantData, t)

	// Length of record
	AppendInterface(uint64(pp.Size+16), wantData, t)

	// Payload
	AppendInterface(pp.Stream.(*bytes.Buffer).Bytes(), wantData, t)

	wantBytes := wantData.Bytes()

	actual := bytes.NewBuffer([]byte{})

	pos := uint64(0)

	num, err := pp.Write(actual, &pos)
	if err != nil {
		t.Errorf("an error occured while writing the entry %v: %e", pp, err)
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

func TestStreamPayload(t *testing.T) {
	pe := GetPayloadTest()
	pe2 := GetPayloadTest()

	buf := bytes.NewBuffer([]byte{})
	bufpos := uint64(0)

	streambuf := bytes.NewBuffer([]byte{})
	streambufpos := uint64(0)

	pe.Write(buf, &bufpos)
	pe2.WriteStream(streambuf, &streambufpos)

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

func TestPayloadChan(t *testing.T) {
	pe := GetPayloadTest()
	pe2 := GetPayloadTest()

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
					t.Errorf("an error occured while writing the entry %v: %e", pe, don)
				}
				break L
			default:
				continue
			}
		}
	}()

	pe2.WriteChannel(ch, &chanpos)
	done <- nil

	bufBytes := buf.Bytes()

	for i := range bufBytes {
		if bufBytes[i] != chanres[i] {
			t.Errorf("mismatch at position %d, expected %b but got %b", i, bufBytes[i], chanres[i])
		}
	}

	close(ch)
}
