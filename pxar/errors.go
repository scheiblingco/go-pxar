package pxar

import "fmt"

type PxarPayloadSizeError struct {
	Expected uint64
	Actual   uint64
}

func (e *PxarPayloadSizeError) Error() string {
	return fmt.Sprintf("payload size mismatch (expected %d, got %d)", e.Expected, e.Actual)
}
