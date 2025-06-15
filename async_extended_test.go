package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
)

// BenchmarkSyncVsAsync compares performance of sync vs async archive creation
func BenchmarkSyncVsAsync(b *testing.B) {
	// Create test archive
	pa := PBSArchive{
		Filename: "benchmark.pxar",
	}
	pa.AddFolder("./test-enc")

	b.Run("Sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer([]byte{})
			err := pa.ToBuffer(buf)
			if err != nil {
				b.Fatalf("Failed to create sync archive: %v", err)
			}
		}
	})

	b.Run("Async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer([]byte{})
			err := pa.ToBufferAsync(buf)
			if err != nil {
				b.Fatalf("Failed to create async archive: %v", err)
			}
		}
	})
}

// TestAsyncConfiguration tests different async configuration options
func TestAsyncConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		asyncMode    bool
		asyncWorkers int
		expectError  bool
	}{
		{"DefaultSync", false, 0, false},
		{"AsyncDefault", true, 0, false},
		{"Async1Worker", true, 1, false},
		{"Async4Workers", true, 4, false},
		{"Async8Workers", true, 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := PBSArchive{
				Filename:     "test_config.pxar",
				AsyncMode:    tt.asyncMode,
				AsyncWorkers: tt.asyncWorkers,
			}
			pa.AddFolder("./test-enc")

			buf := bytes.NewBuffer([]byte{})
			var err error
			
			if tt.asyncMode {
				err = pa.ToBufferAsync(buf)
			} else {
				err = pa.ToBuffer(buf)
			}

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				t.Logf("Archive created successfully with %d bytes", buf.Len())
			}
		})
	}
}

// TestAsyncConsistency verifies that async operations are deterministic
func TestAsyncConsistency(t *testing.T) {
	pa := PBSArchive{
		Filename:     "consistency.pxar",
		AsyncWorkers: 2,
	}
	pa.AddFolder("./test-enc")

	// Create multiple async archives
	var archives [][]byte
	for i := 0; i < 5; i++ {
		buf := bytes.NewBuffer([]byte{})
		err := pa.ToBufferAsync(buf)
		if err != nil {
			t.Fatalf("Failed to create async archive %d: %v", i, err)
		}
		archives = append(archives, buf.Bytes())
	}

	// Verify all archives are identical
	reference := archives[0]
	for i, archive := range archives[1:] {
		if !bytes.Equal(reference, archive) {
			t.Errorf("Archive %d differs from reference. Reference len: %d, Archive len: %d", 
				i+1, len(reference), len(archive))
		}
	}

	t.Logf("All 5 async archives are identical (%d bytes)", len(reference))
}

// TestAsyncWithLargeFiles tests async performance with larger test data
func TestAsyncWithLargeFiles(t *testing.T) {
	// Create a temporary directory with larger files
	tempDir := "/tmp/go-pxar-large-test"
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some larger test files
	testData := make([]byte, 1024*1024) // 1MB of data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("%s/largefile%d.dat", tempDir, i)
		err := os.WriteFile(filename, testData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	pa := PBSArchive{
		Filename:     "large_test.pxar",
		AsyncWorkers: 4,
	}
	pa.AddFolder(tempDir)

	// Test sync version
	syncStart := time.Now()
	syncBuf := bytes.NewBuffer([]byte{})
	err = pa.ToBuffer(syncBuf)
	if err != nil {
		t.Fatalf("Failed to create sync archive: %v", err)
	}
	syncDuration := time.Since(syncStart)

	// Test async version
	asyncStart := time.Now()
	asyncBuf := bytes.NewBuffer([]byte{})
	err = pa.ToBufferAsync(asyncBuf)
	if err != nil {
		t.Fatalf("Failed to create async archive: %v", err)
	}
	asyncDuration := time.Since(asyncStart)

	// Verify results are identical
	if !bytes.Equal(syncBuf.Bytes(), asyncBuf.Bytes()) {
		t.Errorf("Sync and async archives differ. Sync: %d bytes, Async: %d bytes",
			syncBuf.Len(), asyncBuf.Len())
	}

	t.Logf("Large file test completed successfully")
	t.Logf("Sync duration: %v, Async duration: %v", syncDuration, asyncDuration)
	t.Logf("Archive size: %d bytes", syncBuf.Len())
}