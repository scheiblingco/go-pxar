package nodes

import (
	"bytes"
	"os"

	"github.com/dchest/siphash"
	"github.com/scheiblingco/go-pxar/pxar"
)

func (ref *SymlinkRef) GetChildren() []NodeRef {
	return []NodeRef{}
}

func (ref *SymlinkRef) GetHash() uint64 {
	return siphash.Hash(pxar.PXAR_HASH_KEY_1, pxar.PXAR_HASH_KEY_2, []byte(ref.Name))
}

func (ref *SymlinkRef) WriteCatalogue(buf *bytes.Buffer, pos *uint64, parentStartPos uint64) ([]byte, uint64, error) {
	sBuf := bytes.NewBuffer([]byte{})

	sBuf.WriteByte(byte(pxar.SymlinkEntry))
	written := 1

	filenameLen := MakeUvarint(uint64(len(ref.Name)))
	n, err := sBuf.Write(filenameLen)
	if err != nil {
		return nil, 0, err
	}
	written += n

	n, err = sBuf.WriteString(ref.Name)
	if err != nil {
		return nil, 0, err
	}
	written += n

	return sBuf.Bytes(), uint64(written), nil
}

func (ref *SymlinkRef) WritePayload(buf *bytes.Buffer, pos *uint64) (uint64, error) {
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

	// Get symlink target
	target, err := os.Readlink(ref.AbsPath)
	if err != nil {
		return 0, err
	}

	symlink := pxar.PxarSymlink{
		Target: target,
	}

	_, err = symlink.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	return *pos - startPos, nil
}
