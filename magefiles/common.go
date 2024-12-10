package main

import (
	"log/slog"
	"os"

	_ "github.com/magefile/mage/sh"
)

var workingDir string

func init() {
	logLevel := new(slog.LevelVar)

	if val, set := os.LookupEnv("MAGE_LOG_LEVEL"); set {
		_ = logLevel.UnmarshalText([]byte(val))
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})

	slog.SetDefault(slog.New(handler))

	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		workingDir = wd
	}
}
