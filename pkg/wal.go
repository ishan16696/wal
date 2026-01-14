package wal

import (
	"bufio"
	"context"
	"encoding/binary"
	"hash/crc32"
	"io"
	pb "main/proto"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	// DefaultSegmentSizeBytes is the default preallocated size of each wal segment file.
	DefaultSegmentSizeBytes int64 = 64 * 1000 * 1000 // 64MB
	//DefaultWALSegments is a default no. of wal segments need to be maintained.
	DefaultWALSegments int8 = 5
	// DefaultWaitBeforeSyncTime is a default time before flushing the data into disk
	DefaultWaitBeforeSyncTime = 100 * time.Millisecond
)

var _ WAL = (*wal)(nil)

type Action interface {
	Do() error
}

type ActionFunc func() error

func (f ActionFunc) Do() error {
	return f()
}

func newSyncWatchDog(syncTimer time.Ticker, actionFunc Action, cancel context.CancelFunc) syncWatchDog {
	return &walSync{
		syncInterval: syncTimer,
		syncCancel:   cancel,
		action:       actionFunc,
	}
}

// OpenWAL creates or opens a WAL at the given path.
// 1. It creates directory path (if doesn't exist).
// 2. If wal files already exist then it open the last wal segments.
func OpenWAL(lg *zap.Logger, dirpath string) (WAL, error) {
	if err := os.MkdirAll(dirpath, 0600); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}

	var currentWALSegmentNumber int
	var fd *os.File
	if len(files) == 0 {
		//  wal segment doesn't already exist.
		// create the first wal segment
		fd, err = createWALSegment(dirpath, 0)
		if err != nil {
			return nil, err
		}
		currentWALSegmentNumber = 0

		if err := fd.Close(); err != nil {
			return nil, err
		}
	} else {
		// get the last wal file segment number.
		currentWALSegmentNumber, err = getLastWalSegmentNumber(files)
		if err != nil {
			return nil, err
		}
	}

	fd, err = os.OpenFile(filepath.Join(dirpath+getWalSegmentName(currentWALSegmentNumber)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	// Seek to the end of the file
	if _, err = fd.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	wal := &wal{
		dataDir:           dirpath,
		log:               lg,
		maxWALSegments:    DefaultWALSegments,
		waitToSync:        DefaultWaitBeforeSyncTime,
		maxWALSegmentSize: DefaultSegmentSizeBytes,
		currentSegmentFd:  fd,
		shouldNotSync:     false,
		bufWriter:         bufio.NewWriter(fd),
		lastSequenceNo:    0,
	}

	wal.lastSequenceNo, err = wal.getLastSequenceNo()
	if err != nil {
		lg.Error(
			"failed to read last sequence number from wal segment",
			zap.Error(err),
		)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	syncAction := ActionFunc(func() error {
		// sync to disk
		wal.lock.Lock()
		defer wal.lock.Unlock()
		if err := wal.syncToDisk(); err != nil {
			lg.Error("unable to flush to the disk", zap.Error(err))
			return err
		}
		return nil
	})
	syncWatcher := newSyncWatchDog(*time.NewTicker(DefaultWaitBeforeSyncTime), syncAction, cancel)
	syncWatcher.Start(ctx)

	return wal, nil
}

func (w *walSync) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-w.syncInterval.C:
				// sync the file to disk
				w.action.Do()

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (w *walSync) Stop() {
	w.syncCancel()
}

// syncToDisk flushs any data in the WAL's in-memory buffer to the segment file.
// It also calls fsync on the segment file(if fsync isn't disabled).
func (w *wal) syncToDisk() error {
	if err := w.bufWriter.Flush(); err != nil {
		return err
	}

	if !w.shouldNotSync {
		if err := w.currentSegmentFd.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (w *wal) Append(record Record) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.lastSequenceNo++

	WAL_Record := &pb.WAL_Record{
		LogSequenceNumber: w.lastSequenceNo,
		CRC:               crc32.ChecksumIEEE(append(record, byte(w.lastSequenceNo))),
		Payload:           record,
	}

	return w.writeToBuffer(WAL_Record)
}

func (w *wal) writeToBuffer(walRecord *pb.WAL_Record) error {
	marsheled_wal_record, err := proto.Marshal(walRecord)
	if err != nil {
		w.log.Fatal("failed to marshal the wal record", zap.Error(err))
		return err
	}

	size_of_wal_record := int32(len(marsheled_wal_record))
	if err := binary.Write(w.bufWriter, binary.LittleEndian, size_of_wal_record); err != nil {
		return err
	}
	return nil
}

func (w *wal) ReadAll() ([]Record, error) {
	// TODO:
	// 1. Seek to start of file
	// 2. Loop: read [length prefix], then [record]
	// 3. Validate checksum (if implemented)
	// 4. Collect valid records, stop at EOF

	allRecords := make([]Record, 0)
	files, err := os.ReadDir(w.dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
	}

	return allRecords, nil
}

func (w *wal) LogSeqmentation() error {
	// TODO:
	// 1. You may implement by creating a new empty file
	// 2. Or support segment rotation (rename old file, create new one)

	files, err := os.ReadDir(w.dataDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		// no log segmentation is required
		return nil
	}

	noOfWALSegments, err := getLastWalSegmentNumber(files)
	if err != nil {
		return err
	}

	if w.maxWALSegments == int8(noOfWALSegments)+1 {
		// 1. Closed the current wal segments.
		// 2. Delete the old wal segments.
		// 3. Create a new wal segments.

	} else {
		// noOfWALSegments < maxWALSegments
		// 1. Closed the current wal segments.
		// 2. Create a new wal segments.

	}

	return nil
}

func (w *wal) Close() error {
	return w.currentSegmentFd.Close()
}

func (w *wal) getLastSequenceNo() (uint32, error) {
	wal_entry, err := w.getLastWALEntry()
	if err != nil {
		return 0, err
	}

	if wal_entry != nil {
		return wal_entry.GetLogSequenceNumber(), nil
	}
	return 0, nil
}

func (w *wal) getLastWALEntry() (*pb.WAL_Record, error) {
	return &pb.WAL_Record{}, nil
}
