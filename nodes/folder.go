package nodes

import (
	"bytes"
	"fmt"

	"github.com/dchest/siphash"
	"github.com/scheiblingco/go-pxar/pxar"
)

func (ref *FolderRef) GetChildren() []NodeRef {
	return ref.Children
}

func (ref *FolderRef) GetHash() uint64 {
	return siphash.Hash(pxar.PXAR_HASH_KEY_1, pxar.PXAR_HASH_KEY_2, []byte(ref.Name))
}

func (ref *FolderRef) WriteCatalogue(buf *bytes.Buffer, pos *uint64, parentStartPos uint64) ([]byte, uint64, error) {
	startPos := *pos
	totalLen := uint64(0)
	// Add folder length
	table := bytes.NewBuffer([]byte{})

	lenUvi := MakeUvarint(uint64(len(ref.Children)))

	n, err := table.Write(lenUvi)
	if err != nil {
		return nil, 0, err
	}

	totalLen += uint64(n)
	*pos += uint64(n)

	for _, child := range ref.Children {
		chBytes, n, err := child.WriteCatalogue(buf, pos, startPos)
		if err != nil {
			return nil, 0, err
		}

		table.Write(chBytes)
		totalLen += uint64(len(chBytes))
		*pos += n + uint64(len(chBytes))
	}

	buf.Write(MakeUvarint(uint64(table.Len())))
	buf.Write(table.Bytes())

	selfItem := bytes.NewBuffer([]byte{})
	selfItem.WriteByte(byte(pxar.DirectoryEntry))
	selfItem.Write(MakeUvarint(uint64(len(ref.Name))))
	selfItem.WriteString(ref.Name)
	selfItem.Write(MakeUvarint(uint64(totalLen + 1)))

	return selfItem.Bytes(), totalLen, nil
}

func (ref *FolderRef) WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	startPos := *pos

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.Write(buf, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	for _, child := range ref.Children {
		n, err := child.WritePayload(buf, pos)
		if err != nil {
			return 0, err
		}

		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   child.GetHash(),
			Offset: *pos,
			Length: n,
		})
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}

func (ref *FolderRef) WritePayloadChannel(ch chan []byte, pos *uint64) (uint64, error) {
	startPos := *pos

	if len(ref.GoodbyeItems) > 0 {
		// Reset goodbye items in case file is written more than once
		ref.GoodbyeItems = []pxar.GoodbyeItem{}
	}

	if !ref.IsRoot {
		filename := pxar.PxarFilename{
			Content: ref.Name,
		}

		n, err := filename.WriteChannel(ch, pos)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Filename %s is %d bytes (+ entry bytes 56)\r\n", ref.Name, n)
	}

	folderStart := *pos

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := entry.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	for _, child := range ref.Children {
		n, err := child.WritePayloadChannel(ch, pos)
		if err != nil {
			return 0, err
		}

		ref.GoodbyeItems = append(ref.GoodbyeItems, pxar.GoodbyeItem{
			Hash:   child.GetHash(),
			Offset: *pos,
			Length: n,
		})
	}

	gbi := pxar.PxarGoodbye{
		Items:        ref.GoodbyeItems,
		FolderStart:  folderStart,
		GoodbyeStart: *pos,
	}

	_, err = gbi.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	if ref.IsRoot {
		// Write special catalogue pointer
		fmt.Println("Root")
	}

	return *pos - startPos, nil
}
