package pxar

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PxarEntry struct {
	Mode         uint64
	Flags        uint64
	Uid          uint32
	Gid          uint32
	MtimeSecs    uint64
	MtimeNanos   uint32
	MtimePadding uint32
}

func (p *PxarEntry) Write(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_ENTRY,
		Length:    56,
	}

	err := binary.Write(buf, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	err = binary.Write(buf, binary.LittleEndian, p)
	if err != nil {
		return 0, err
	}

	*pos += pd.Length

	return pd.Length, nil
}

func (p *PxarEntry) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_ENTRY,
		Length:    56,
	}

	err := binary.Write(stream, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	err = binary.Write(stream, binary.LittleEndian, p)
	if err != nil {
		return 0, err
	}

	*pos += pd.Length

	return pd.Length, nil
}
