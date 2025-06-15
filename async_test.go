package main

import (
	"bytes"
	"testing"
)

// TestAsyncVsSync verifies that async and sync archive creation produce identical results
func TestAsyncVsSync(t *testing.T) {
	// Create a new PXAR archive
	pa := PBSArchive{
		Filename: "test_async.pxar",
	}

	// Add test folder
	pa.AddFolder("./test-enc")

	// Create synchronous archive
	syncBuf := bytes.NewBuffer([]byte{})
	err := pa.ToBuffer(syncBuf)
	if err != nil {
		t.Fatalf("Failed to create synchronous archive: %v", err)
	}

	// Create asynchronous archive
	asyncBuf := bytes.NewBuffer([]byte{})
	err = pa.ToBufferAsync(asyncBuf)
	if err != nil {
		t.Fatalf("Failed to create asynchronous archive: %v", err)
	}

	// Compare the results
	if !bytes.Equal(syncBuf.Bytes(), asyncBuf.Bytes()) {
		t.Errorf("Async and sync archives differ in content. Sync len: %d, Async len: %d", 
			syncBuf.Len(), asyncBuf.Len())
	}

	t.Logf("Success: Both sync and async archives are identical (%d bytes)", syncBuf.Len())
}

// TestAsyncWithWorkers tests async functionality with different worker counts
func TestAsyncWithWorkers(t *testing.T) {
	// Create a new PXAR archive
	pa := PBSArchive{
		Filename:     "test_workers.pxar",
		AsyncWorkers: 2, // Use 2 workers
	}

	// Add test folder
	pa.AddFolder("./test-enc")

	// Create synchronous archive for reference
	syncBuf := bytes.NewBuffer([]byte{})
	err := pa.ToBuffer(syncBuf)
	if err != nil {
		t.Fatalf("Failed to create synchronous archive: %v", err)
	}

	// Create asynchronous archive with 2 workers
	asyncBuf := bytes.NewBuffer([]byte{})
	err = pa.ToBufferAsync(asyncBuf)
	if err != nil {
		t.Fatalf("Failed to create asynchronous archive with 2 workers: %v", err)
	}

	// Compare the results
	if !bytes.Equal(syncBuf.Bytes(), asyncBuf.Bytes()) {
		t.Errorf("Async (2 workers) and sync archives differ in content. Sync len: %d, Async len: %d", 
			syncBuf.Len(), asyncBuf.Len())
	}

	t.Logf("Success: Both sync and async (2 workers) archives are identical (%d bytes)", syncBuf.Len())
}

// TestAsyncChannel tests async functionality with channels
func TestAsyncChannel(t *testing.T) {
	// Create a new PXAR archive
	pa := PBSArchive{
		Filename: "test_async_channel.pxar",
	}

	// Add test folder
	pa.AddFolder("./test-enc")

	// Create synchronous archive for reference
	syncBuf := bytes.NewBuffer([]byte{})
	err := pa.ToBuffer(syncBuf)
	if err != nil {
		t.Fatalf("Failed to create synchronous archive: %v", err)
	}

	// Create asynchronous archive with channel
	ch := make(chan []byte, 100)
	done := make(chan error, 1)
	var asyncData []byte

	// Start goroutine to collect channel data
	go func() {
		for data := range ch {
			asyncData = append(asyncData, data...)
		}
		done <- nil
	}()

	// Create async archive via channel
	err = pa.ToChannelAsync(ch)
	if err != nil {
		t.Fatalf("Failed to create asynchronous channel archive: %v", err)
	}
	
	close(ch)
	<-done

	// Compare the results
	if !bytes.Equal(syncBuf.Bytes(), asyncData) {
		t.Errorf("Async channel and sync archives differ in content. Sync len: %d, Async len: %d", 
			syncBuf.Len(), len(asyncData))
	}

	t.Logf("Success: Both sync and async channel archives are identical (%d bytes)", syncBuf.Len())
}