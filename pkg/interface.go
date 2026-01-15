package wal

import "context"

// Record represents a single WAL entry.
type Record []byte

// WAL is the interface for Write-Ahead Log functionality.
type WAL interface {
	// Append a record to the WAL and ensure durability.
	Append(record Record) error

	// ReadAll reads all valid records from the WAL (for recovery).
	ReadAll() ([]Record, error)

	// LogSeqmentation does the log segmentation of wal file.
	LogSeqmentation() error

	// Close WAL file.
	Close() error
}

// syncWatchDog manages a goroutine that can be started and stopped.
type syncWatchDog interface {
	// Start starts a goroutine using the given context.
	Start(ctx context.Context)
	// Stop stops the goroutine started by Start.
	Stop()
}
