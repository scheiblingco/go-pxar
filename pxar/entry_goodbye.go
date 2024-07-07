package pxar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type GoodbyeItem struct {
	Hash   uint64
	Offset uint64
	Length uint64
}

type PxarGoodbye struct {
	Items        []GoodbyeItem
	FolderStart  uint64
	GoodbyeStart uint64
}

func (p *PxarGoodbye) Write(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	fmt.Println("Starting goodbye at", *pos)
	pd := PxarDescriptor{
		EntryType: PXAR_GOODBYE,
		Length: 40 + uint64(
			24*(len(p.Items)),
		),
	}

	err := binary.Write(buf, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	*pos += 16

	binTree := make([]GoodbyeItem, len(p.Items))
	GetBinaryHeap(p.Items, &binTree)

	for _, item := range binTree {
		item.Offset = p.GoodbyeStart - (item.Offset - item.Length)
		binary.Write(buf, binary.LittleEndian, item)
		*pos += 24
	}

	final := &GoodbyeItem{
		Offset: p.GoodbyeStart - p.FolderStart,
		Length: pd.Length,
		Hash:   PXAR_GOODBYE_TAIL_MARKER,
	}

	binary.Write(buf, binary.LittleEndian, final)
	*pos += 24

	fmt.Println("Finishing goodbye at ", *pos)
	return pd.Length, nil
}

func (p *PxarGoodbye) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_GOODBYE,
		Length: 40 + uint64(
			24*(len(p.Items)),
		),
	}

	err := binary.Write(stream, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	*pos += 16

	binTree := make([]GoodbyeItem, len(p.Items))
	GetBinaryHeap(p.Items, &binTree)

	for _, item := range binTree {
		item.Offset = p.GoodbyeStart - (item.Offset - item.Length)
		binary.Write(stream, binary.LittleEndian, item)
		*pos += 24
	}

	final := &GoodbyeItem{
		Offset: p.GoodbyeStart - p.FolderStart,
		Length: pd.Length,
		Hash:   PXAR_GOODBYE_TAIL_MARKER,
	}

	binary.Write(stream, binary.LittleEndian, final)
	*pos += 24

	return pd.Length, nil
}

func (p *PxarGoodbye) WriteChannel(ch chan []byte, pos *uint64) (uint64, error) {
	buf := bytes.NewBuffer([]byte{})
	n, err := p.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	ch <- buf.Bytes()

	return n, nil
}
