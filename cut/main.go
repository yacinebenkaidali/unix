package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Config struct {
	delim   rune
	files   []string
	columns []int
}

var doneCh = make(chan struct{})
var results = make(chan string, 100)
var errorCh = make(chan error)
var wg sync.WaitGroup

func main() {
	fields := flag.String("f", "", "Column name(s)")
	d := flag.String("d", "\t", "Delimiter by which we separate the lines")
	flag.Parse()
	files := flag.Args()

	columns, err := ConvertStringsToInts(*fields)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)

	}
	cfg := Config{
		delim:   GetSeparator(*d),
		files:   files,
		columns: columns,
	}

	go run(os.Stdout, cfg)

	<-doneCh
	<-doneCh
	<-doneCh
	close(doneCh)
}

func run(out io.Writer, cfg Config) {

	go func() {
		for content := range results {
			fmt.Fprintln(os.Stdout, content)
			fmt.Fprintln(os.Stdout, "=========/==========")
		}
		doneCh <- struct{}{}
	}()

	go func() {
		for err := range errorCh {
			fmt.Fprintln(os.Stderr, err)
		}
		doneCh <- struct{}{}
	}()

	input, cleanup := getInput(cfg.files)
	if cleanup != nil {
		defer cleanup()
	}

	for _, in := range input {
		wg.Add(1)
		go cutFileContent(in, cfg.columns, cfg.delim)
	}

	wg.Wait()
	close(results)
	close(errorCh)
	doneCh <- struct{}{}
}

func cutFileContent(in io.Reader, columns []int, delim rune) {
	defer wg.Done()
	var content strings.Builder
	csvReader := csv.NewReader(in)
	csvReader.Comma = delim

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCh <- err
		}
		columnsContent := make([]string, 0, len(columns))
		for _, c := range columns {
			if c < len(line) {
				columnsContent = append(columnsContent, line[c-1])
			}
		}
		content.WriteString(fmt.Sprintf("%s\n", strings.Join(columnsContent, string(delim))))
	}
	results <- content.String()
}

func getInput(files []string) ([]io.Reader, func()) {
	stdInfo, err := os.Stdin.Stat()
	if err != nil {
		errorCh <- err
	}
	if stdInfo.Size() > 0 {
		return []io.Reader{os.Stdin}, nil
	}

	if len(files) > 0 {
		filesReaders := make([]io.Reader, 0, len(files))
		filesPtrs := make([]*os.File, 0, len(files))

		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				errorCh <- err
				continue
			}
			filesReaders = append(filesReaders, f)
			filesPtrs = append(filesPtrs, f)
		}
		return filesReaders, func() {
			for _, file := range filesPtrs {
				file.Close()
			}
		}
	}
	return []io.Reader{}, nil
}
