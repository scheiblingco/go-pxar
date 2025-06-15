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

func (ref *FileRef) WriteCatalogue(buf *bytes.Buffer, pos *uint64, parentStartPos uint64) ([]byte, uint64, error) {
	sBuf := bytes.NewBuffer([]byte{})
	sBuf.WriteByte(byte(pxar.FileEntry))

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

	fileSize := MakeUvarint(uint64(ref.Stat.Size))
	n, err = sBuf.Write(fileSize)
	if err != nil {
		return nil, 0, err
	}
	written += n

	mTime := MakeUvarint(ref.Stat.MtimeSecs)
	n, err = sBuf.Write(mTime)
	if err != nil {
		return nil, 0, err
	}
	written += n

	return sBuf.Bytes(), uint64(written), nil
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

func (ref *FileRef) WritePayloadChannel(ch chan []byte, pos *uint64) (uint64, error) {
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

	_, err := filename.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	_, err = entry.WriteChannel(ch, pos)
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

	_, err = payload.WriteChannel(ch, pos)
	if err != nil {
		return 0, err
	}

	return *pos - startPos, nil
}

// WritePayloadAsync implements concurrent file processing - for files, it's the same as synchronous
// since individual file writes must be sequential, but this maintains interface compatibility
func (ref *FileRef) WritePayloadAsync(buf *bytes.Buffer, pos *uint64, workers int) (uint64, error) {
	// For individual files, async doesn't provide benefits since file content must be written sequentially
	// The async benefit comes at the directory level where multiple files can be processed concurrently
	return ref.WritePayload(buf, pos)
}

// WritePayloadChannelAsync implements concurrent file processing - for files, it's the same as synchronous  
func (ref *FileRef) WritePayloadChannelAsync(ch chan []byte, pos *uint64, workers int) (uint64, error) {
	// For individual files, async doesn't provide benefits since file content must be written sequentially
	// The async benefit comes at the directory level where multiple files can be processed concurrently
	return ref.WritePayloadChannel(ch, pos)
}
