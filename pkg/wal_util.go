package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getWalSegmentName(walSegmentNumber int) string {
	return fmt.Sprintf("%d-segment.wal", walSegmentNumber)
}

func createWALSegment(dirpath string, walSegmentNumber int) (*os.File, error) {
	fd, err := os.Create(filepath.Join(dirpath + getWalSegmentName(walSegmentNumber)))
	if err != nil {
		return nil, err
	}
	return fd, nil
}

// getLastWalSegmentFile returns the last wal segment file.
func getLastWalSegmentFile(files []os.DirEntry) os.DirEntry {
	var lastFile os.DirEntry

	// files are coming in ascending order by their name.
	for _, file := range files {
		lastFile = file
	}
	return lastFile
}

func getLastWalSegmentNumber(files []os.DirEntry) (int, error) {
	lastWALfile := getLastWalSegmentFile(files)

	before, _, found := strings.Cut(lastWALfile.Name(), "-")
	if !found {
		return 0, fmt.Errorf("wal segments aren't named correctly")
	}

	segmentNumber, err := strconv.Atoi(before)
	if err != nil {
		return 0, err
	}

	return segmentNumber, nil
}

// func unmarshalWALEntry(data []byte, entry *pb.WAL_Record) {
// 	if err := proto.Unmarshal(data, entry); err != nil {

// 	}
// }
