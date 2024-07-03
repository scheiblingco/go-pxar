package pxar

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PxarFilename struct {
	Content string
}

func (p *PxarFilename) Write(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_FILENAME,
		Length:    17 + uint64(len(p.Content)),
	}

	err := binary.Write(buf, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	buf.WriteString(p.Content)
	buf.WriteByte(0x00)

	*pos += pd.Length

	return pd.Length, nil
}

func (p *PxarFilename) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_FILENAME,
		Length:    17 + uint64(len(p.Content)),
	}

	err := binary.Write(stream, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	_, err = stream.Write([]byte(p.Content))
	if err != nil {
		return 0, err
	}

	_, err = stream.Write([]byte{0x00})
	if err != nil {
		return 0, err
	}

	*pos += pd.Length

	return pd.Length, nil
}
