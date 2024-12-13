package main

import (
	"embed"
	"log/slog"
	"os"
	"text/template"

	_ "github.com/magefile/mage/sh"
)

var (
	workingDir string
	//go:embed templates/*.tmpl
	templatesFS embed.FS
	templates   *template.Template
)

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

	templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
}
