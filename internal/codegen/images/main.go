package main

import (
	_ "embed"
	"flag"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	//go:embed images.go.tmpl
	rawTemplate    string
	imagesTemplate *template.Template
	flags          *flag.FlagSet
	sourceFilePath string
	outFilePath    string
	envoySpec      struct {
		Repo string
		Tag  string
	}
)

func init() {
	flags = flag.NewFlagSet("images", flag.PanicOnError)

	flags.StringVar(&sourceFilePath, "source-file", "", "Path to the source file")
	flags.StringVar(&outFilePath, "out-file", "", "Path to the output file")
	flags.StringVar(&envoySpec.Repo, "envoy-repo", "", "Container repo for the envoy image")
	flags.StringVar(&envoySpec.Tag, "envoy-tag", "", "Tag for envoy container image")

	var err error
	imagesTemplate, err = template.New("images.go.tmpl").Parse(rawTemplate)
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		slog.Error("Error parsing flags", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srcFile, err := os.Open(sourceFilePath)
	if err != nil {
		slog.Error("Error opening file:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		if err := srcFile.Close(); err != nil {
			slog.Error("Error closing compose file", slog.String("error", err.Error()))
		}
	}()

	outFile, err := os.Create(outFilePath)
	if err != nil {
		slog.Error("Error creating output file:", err)
		os.Exit(1)
	}

	defer func() {
		if err := outFile.Close(); err != nil {
			slog.Error("Error closing output file", slog.String("error", err.Error()))
		}
	}()

	var composeFile struct {
		Services map[string]struct {
			Image string `yaml:"image"`
		}
	}

	if err := yaml.NewDecoder(srcFile).Decode(&composeFile); err != nil {
		slog.Error("Failed to decode compose file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	type imageRef struct {
		Repository string
		Tag        string
	}

	serviceMappings := map[string]string{
		"auth":      "Gotrue",
		"functions": "EdgeRuntime",
		"imgproxy":  "ImgProxy",
		"meta":      "PostgresMeta",
		"realtime":  "Realtime",
		"rest":      "Postgrest",
		"storage":   "Storage",
		"studio":    "Studio",
	}

	templateData := struct {
		Images map[string]imageRef
		Year   int
		Author string
	}{
		Images: make(map[string]imageRef),
		Year:   time.Now().Year(),
		Author: "Peter Kurfer",
	}

	for name, service := range composeFile.Services {
		splitIdx := strings.LastIndex(service.Image, ":")
		repo := service.Image[:splitIdx]
		tag := service.Image[splitIdx+1:]

		mapping, ok := serviceMappings[name]
		if !ok {
			continue
		}

		templateData.Images[mapping] = imageRef{
			Repository: repo,
			Tag:        tag,
		}
	}

	templateData.Images["Envoy"] = imageRef{
		Repository: envoySpec.Repo,
		Tag:        envoySpec.Tag,
	}

	if err := imagesTemplate.ExecuteTemplate(outFile, "images.go.tmpl", templateData); err != nil {
		slog.Error("Failed to execute template", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := outFile.Sync(); err != nil {
		slog.Error("Failed to sync output file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
