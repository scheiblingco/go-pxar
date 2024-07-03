package pxar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
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
	px := ""

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

	sort.Slice(p.Items, func(i, j int) bool {
		return p.Items[i].Hash < p.Items[j].Hash
	})

	binTree := make([]GoodbyeItem, len(p.Items))
	MakeBinaryTree(p.Items, &binTree)

	shouldBe := []byte("U\x15>u[\xed^\xef\xb0\x00\x00\x00\x00\x00\x00\x00@\x00\x00\x00\x00\x00\x00\x00")
	shouldBeItem := GoodbyeItem{}
	binary.Read(bytes.NewReader(shouldBe), binary.LittleEndian, &shouldBeItem)

	fmt.Println(shouldBe)

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

	px = buf.String()
	pxb := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, final)
	binary.Write(pxb, binary.LittleEndian, final)
	px = pxb.String()
	*pos += 24

	fmt.Println(px)

	return pd.Length, nil
}

func (p *PxarGoodbye) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	entryPos := *pos

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

	sort.Slice(p.Items, func(i, j int) bool {
		return p.Items[i].Hash < p.Items[j].Hash
	})

	binTree := make([]GoodbyeItem, len(p.Items))
	MakeBinaryTree(p.Items, &binTree)

	for _, item := range binTree {
		item.Offset = *pos - item.Offset - item.Length
		binary.Write(stream, binary.LittleEndian, item)
		*pos += 24
	}

	final := &GoodbyeItem{
		Offset: *pos,
		Length: pd.Length,
		Hash:   PXAR_GOODBYE_TAIL_MARKER,
	}

	binary.Write(stream, binary.LittleEndian, final)

	*pos += 24

	if entryPos != (*pos - pd.Length) {
		return 0, &PxarPayloadSizeError{
			Expected: pd.Length,
			Actual:   *pos - entryPos,
		}
	}

	return pd.Length, nil
}