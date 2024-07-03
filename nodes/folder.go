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
