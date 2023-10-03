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

type Result struct {
	file   string
	result string
	err    error
}

var doneCh = make(chan struct{})
var resultsCh = make(chan Result, 10)
var filesCh = make(chan string, 10)
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
	close(doneCh)
}

func run(out io.Writer, cfg Config) {
	go sendData(cfg.files)
	wg.Add(1)

	go func() {
		defer wg.Done()
		for file := range filesCh {
			wg.Add(1)
			go cutFileContent(file, cfg.columns, cfg.delim)
		}
	}()
	go handleResults(out)

	wg.Wait()
	close(resultsCh)
	doneCh <- struct{}{}
}

func sendData(files []string) {
	for _, file := range files {
		filesCh <- file
	}
	close(filesCh)
}

func handleResults(out io.Writer) {
	for content := range resultsCh {
		if content.err != nil {
			fmt.Printf("Error occured : %q\n", content.err)
		} else {
			fmt.Fprintf(out, "File name :%s\n", content.file)
			fmt.Fprintf(out, "%s\n", content.result)
		}
		fmt.Fprintln(out, "=========/==========")
	}
	doneCh <- struct{}{}
}

func cutFileContent(file string, columns []int, delim rune) {
	defer wg.Done()
	var content strings.Builder

	f, err := os.Open(file)
	if err != nil {
		resultsCh <- Result{err: err, file: file}
		return
	}

	csvReader := csv.NewReader(f)
	csvReader.Comma = delim

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			resultsCh <- Result{err: err, file: file}
			return
		}
		tmpStr := ""
		for i, c := range columns {
			if c < len(line) {
				tmpStr += line[c-1]
			}
			if i < len(columns)-1 {
				tmpStr += string(delim)
			}
		}
		content.WriteString(fmt.Sprintf("%s\n", tmpStr))
	}
	resultsCh <- Result{
		result: content.String(),
		file:   file,
	}
}
