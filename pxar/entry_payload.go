package pxar

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PxarPayload struct {
	Size   uint64
	Stream io.Reader
}

func (p *PxarPayload) Write(buf *bytes.Buffer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_PAYLOAD,
		Length:    16 + p.Size,
	}

	err := binary.Write(buf, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(buf, p.Stream)
	if err != nil {
		return 0, err
	}

	if n != int64(p.Size) {
		return 0, &PxarPayloadSizeError{
			Expected: p.Size,
			Actual:   uint64(n),
		}
	}

	*pos += pd.Length

	return pd.Length, nil
}

func (p *PxarPayload) WriteStream(stream io.Writer, pos *uint64) (uint64, error) {
	pd := PxarDescriptor{
		EntryType: PXAR_PAYLOAD,
		Length:    16 + p.Size,
	}

	err := binary.Write(stream, binary.LittleEndian, pd)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(stream, p.Stream)
	if err != nil {
		return 0, err
	}

	if n != int64(p.Size) {
		return 0, &PxarPayloadSizeError{
			Expected: p.Size,
			Actual:   uint64(n),
		}
	}

	*pos += pd.Length

	return pd.Length, nil
}
