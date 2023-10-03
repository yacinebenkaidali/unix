package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

type config struct {
	bytes bool
	lines bool
	words bool
	char  bool
}

type Result struct {
	fileName string
	count    int
}

func main() {
	//flags definition
	c := flag.Bool("c", false, "count the number of bytes")
	l := flag.Bool("l", false, "count the number of lines")
	w := flag.Bool("w", false, "count the number of words")
	m := flag.Bool("m", false, "count the number of chars/runes")
	flag.Parse()
	files := flag.Args()

	//channels
	results := make(chan Result, runtime.NumCPU())
	filesCh := make(chan string, runtime.NumCPU())
	done := make(chan struct{})

	config := config{bytes: *c, lines: *l, words: *w, char: *m}

	var wg sync.WaitGroup

	go func() {
		for _, file := range files {
			filesCh <- file
		}
		close(filesCh)
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for fileName := range filesCh {
				file, err := os.Open(fileName)

				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				results <- Result{fileName: fileName, count: count(config, file)}

			}
			wg.Done()
		}()
	}
	go func() {
		for res := range results {
			fmt.Printf("FileName : %s, count : %d\n", res.fileName, res.count)
		}
		done <- struct{}{}
	}()

	wg.Wait()
	close(results)
	<-done
	fmt.Fprintln(os.Stdout, "Done")
}

func count(c config, reader io.Reader) int {
	scanner := bufio.NewScanner(reader)

	switch {
	case c.bytes:
		{
			scanner.Split(bufio.ScanBytes)
		}
	case c.lines:
		{
			scanner.Split(bufio.ScanLines)
		}
	case c.words:
		{
			scanner.Split(bufio.ScanWords)
		}
	case c.char:
		{
			scanner.Split(bufio.ScanRunes)
		}
	}

	wc := 0
	for scanner.Scan() {
		wc++
	}

	return wc
}
