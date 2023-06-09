package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/tools/godoc/util"
)

type options struct {
	chunkSize       int
	includePatterns []string
	excludePatterns []string
	reportPath      string
	source          string
}

var chunkIndex int
var chunkSizeRemaining int
var chunkFile *os.File

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: go2chatgpt [options] <source> <output_folder>\n")
		flag.PrintDefaults()
	}

	var opts options
	chunkSize := flag.Int("chunksize", 13, "Chunk size in KB (GPT-4 max is 16KB)")
	exclude := flag.String("exclude", "", "Comma-separated list of glob patterns to exclude")
	include := flag.String("include", "", "Comma-separated list of glob patterns to include")
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	opts.chunkSize = *chunkSize * 1024
	opts.source = flag.Arg(0)
	opts.reportPath = flag.Arg(1)

	if *include != "" {
		opts.includePatterns = strings.Split(*include, ",")
	}

	if *exclude != "" {
		opts.excludePatterns = append(strings.Split(*exclude, ","), "**/.git/**")
	} else {
		opts.excludePatterns = []string{"**/.git/**"}
	}

	err := os.MkdirAll(opts.reportPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating report directory: %v\n", err)
		os.Exit(1)
	}

	err = filepath.Walk(opts.source, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skips the chunk files in the report directory. Could probably just skip the report
		// directory entirely, but this is easier for now since relative paths and such.
		if strings.Contains(filePath, filepath.Join(opts.reportPath, "chunk")) {
			return nil
		}

		if !info.IsDir() && shouldProcess(opts, filePath) {
			err = doChunk(opts, filePath)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error processing files: %s", err)
	}

	if chunkFile != nil {
		chunkFile.Close()
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// shouldProcess returns true if the file should be processed, false otherwise.
func shouldProcess(opts options, filePath string) bool {
	for _, pattern := range opts.excludePatterns {
		matched, _ := doublestar.PathMatch(pattern, filePath)
		if matched {
			return false
		}
	}
	if len(opts.includePatterns) > 0 {
		for _, pattern := range opts.includePatterns {
			matched, _ := doublestar.PathMatch(pattern, filePath)
			if matched {
				return true
			}
		}
		return false
	}

	for _, pattern := range opts.excludePatterns {
		matched, _ := doublestar.PathMatch(pattern, filePath)
		if matched {
			return false
		}
	}

	return true
}

// doChunk processes a single file and writes it to the chunk files.
func doChunk(opts options, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if len(data) >= opts.chunkSize {
			return opts.chunkSize, data[0:opts.chunkSize], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	scanner.Split(split)

	needToContinue := false
	for scanner.Scan() {
		chunk := scanner.Bytes()
		if !util.IsText(chunk) {
			fmt.Println("Skipping non-text file:", filePath)
			return nil
		}
		relPath, _ := filepath.Rel(opts.source, filePath)

		for len(chunk) > 0 {
			if chunkSizeRemaining <= 0 {
				if chunkFile != nil {
					chunkFile.Close()
				}
				chunkPath := filepath.Join(opts.reportPath, fmt.Sprintf("chunk%d.txt", chunkIndex))
				chunkFile, err = os.Create(chunkPath)
				if err != nil {
					return err
				}
				chunkIndex++
				chunkSizeRemaining = opts.chunkSize
			}

			if needToContinue {
				chunkFile.WriteString(fmt.Sprintf("\n----CONTINUED FILE: %s----\n", relPath))
				needToContinue = false
			} else {
				chunkFile.WriteString(fmt.Sprintf("\n----BEGIN FILE: %s----\n", relPath))
			}

			bytesToWrite := len(chunk)
			if bytesToWrite > chunkSizeRemaining {
				needToContinue = true
				bytesToWrite = chunkSizeRemaining
			}

			chunkFile.Write(chunk[:bytesToWrite])
			chunk = chunk[bytesToWrite:]
			chunkSizeRemaining -= bytesToWrite

			if needToContinue {
				chunkFile.WriteString(fmt.Sprintf("\n----END PART OF FILE: %s----\n", relPath))
			} else {
				chunkFile.WriteString(fmt.Sprintf("\n----END FILE: %s----\n", relPath))
			}

			if chunkSizeRemaining == 0 {
				chunkFile.Close()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
