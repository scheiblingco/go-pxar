package pxar

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PxarSymlink struct {
	Target string
}

func (p *PxarSymlink) Write(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_SYMLINK,
		Length:    uint64(17 + len(p.Target)),
	}

	err := binary.Write(buf, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	*pos += 16

	_, err = buf.Write([]byte(p.Target))
	if err != nil {
		return 0, err
	}

	*pos += uint64(len(p.Target))

	buf.WriteByte(0x00)
	*pos++

	return pd.Length, nil
}

func (p *PxarSymlink) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	buf := bytes.NewBuffer([]byte{})
	_, err := p.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	_, err = stream.Write(buf.Bytes())
	if err != nil {
		return 0, err
	}

	return uint64(buf.Len()), nil
}

func (p *PxarSymlink) WriteChannel(ch chan []byte, pos *uint64) (uint64, error) {
	buf := bytes.NewBuffer([]byte{})

	n, err := p.Write(buf, pos)
	if err != nil {
		return 0, err
	}

	ch <- buf.Bytes()

	return n, nil
}
