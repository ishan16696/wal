package wal

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// wal is a logical representation of the wal storage.
type wal struct {
	// logger to log the information.
	log *zap.Logger
	// maxWALSegmentSize is maximum size in bytes for each wal segment file size.
	maxWALSegmentSize int64
	// maxWALSegments is a maximum no. of wal segments need to be maintained.
	maxWALSegments int8
	// dataDir is a working directory under which wal file will be stored.
	dataDir string
	// currentSegmentFd is a file descriptor for the current wal segment.
	currentSegmentFd *os.File
	// lock to synchronize the wal writes.
	lock sync.Mutex
	// if set, do not fsync
	shouldNotSync bool
	// lastSequenceNo is used to track the last sequence number of wal segment.
	lastSequenceNo uint32
	// waitToSync is interval before wal flushes into the disk.
	waitToSync time.Duration
	// bufWriter to write the log file.
	bufWriter *bufio.Writer
}

type walSync struct {
	action       Action
	syncInterval time.Ticker
	syncCancel   context.CancelFunc
}
