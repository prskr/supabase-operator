package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

//go:embed migrations_ignore
var ignoredMigrationsRaw []byte

var (
	ignoredMigrationsList        []string
	outputDir                    string
	databaseNameReplacementRegex = regexp.MustCompile(`(?i)(ON DATABASE) postgres`)
)

func init() {
	scanner := bufio.NewScanner(bytes.NewReader(ignoredMigrationsRaw))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" && !strings.HasPrefix(line, "#") {
			ignoredMigrationsList = append(ignoredMigrationsList, line)
		}
	}
}

func main() {
	flag.StringVar(&outputDir, "output-dir", ".", "Output directory for the extracted migrations")
	flag.Parse()

	inputFiles := flag.Args()

	if len(inputFiles) < 1 {
		log.Fatal("no migration input files provided")
	}

	for _, in := range inputFiles {
		fileName := filepath.Base(in)
		if slices.Contains(ignoredMigrationsList, fileName) {
			log.Printf("Skipping migration %s", fileName)
			continue
		}

		migrationOutDir := filepath.Join(outputDir, filepath.Base(filepath.Dir(in)))

		if err := processFile(in, migrationOutDir); err != nil {
			log.Fatalf("Failed to process file %s: %v", in, err)
		}
	}
}

func processFile(in, destDir string) (err error) {
	inFile, err := os.Open(in)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, inFile.Close())
	}()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("failed to create dest dir: %w", err)
	}

	_, fileName := filepath.Split(in)

	outFile, err := os.Create(filepath.Join(destDir, fileName))
	if err != nil {
		return fmt.Errorf("failed to create out file: %w", err)
	}
	defer func() {
		err = errors.Join(err, outFile.Close())
	}()

	_, err = io.Copy(outFile, replacementReader(inFile, databaseNameReplacementRegex, `$1 {{ .DbName }}`))

	return err
}

type readerFunc func(p []byte) (int, error)

func (rf readerFunc) Read(p []byte) (int, error) {
	return rf(p)
}

func replacementReader(r io.Reader, matcher *regexp.Regexp, replacement string) io.Reader {
	buf := bytes.NewBuffer(nil)
	scanner := bufio.NewScanner(r)
	// Use a 1MB buffer for lines to handle large migration files
	const maxScanCapacity = 1024 * 1024
	scanner.Buffer(make([]byte, 0, maxScanCapacity), maxScanCapacity)
	scanner.Split(bufio.ScanLines)

	return readerFunc(func(p []byte) (int, error) {
		if buf.Len() >= len(p) {
			return buf.Read(p)
		}

		for buf.Len() < len(p) && scanner.Scan() {
			line := scanner.Text()
			replaced := matcher.ReplaceAllString(line, replacement)
			buf.WriteString(replaced + "\n")
		}

		if scannErr := scanner.Err(); scannErr != nil {
			return 0, scannErr
		}

		return buf.Read(p)
	})
}
