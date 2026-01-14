package main

import (
	wal "main/pkg"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	_, err = wal.OpenWAL(logger, "")
	if err != nil {
		panic(err)
	}
}
