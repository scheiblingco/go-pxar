package nodes

import (
	"bytes"
	"os"

	"github.com/dchest/siphash"
	"github.com/scheiblingco/go-pxar/pxar"
)

func (ref *FileRef) GetChildren() []NodeRef {
	return []NodeRef{}
}

func (ref *FileRef) GetHash() uint64 {
	return siphash.Hash(pxar.PXAR_HASH_KEY_1, pxar.PXAR_HASH_KEY_2, []byte(ref.Name))
}

func (ref *FileRef) WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	startPos := *pos

	filename := pxar.PxarFilename{
		Content: ref.Name,
	}

	entry := pxar.PxarEntry{
		Mode:         ref.Stat.Mode,
		Uid:          ref.Stat.Uid,
		Gid:          ref.Stat.Gid,
		MtimeSecs:    ref.Stat.MtimeSecs,
		MtimeNanos:   ref.Stat.MtimeNsecs,
		MtimePadding: 0,
	}

	_, err := filename.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	_, err = entry.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	f, err := os.Open(ref.AbsPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	payload := pxar.PxarPayload{
		Size:   uint64(ref.Stat.Size),
		Stream: f,
	}

	_, err = payload.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	return *pos - startPos, nil
}
