package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Config struct {
	lines      int
	bytes      int
	filesNames []string
	resutls    chan Result
}

type Result struct {
	fileName string
	data     string
	error
}

func main() {
	nbLines := flag.Int("n", -1, "Specify the number of lines to be read!")
	nbBytes := flag.Int("c", -1, "Specify the number of bytes to be read!")
	flag.Parse()

	c := Config{
		lines:   *nbLines,
		bytes:   *nbBytes,
		resutls: make(chan Result, len(flag.Args())),
	}

	if len(flag.Args()) > 1 {
		c.filesNames = flag.Args()
	} else {
		c.filesNames = []string{}
	}
	var wg sync.WaitGroup

	doneCh := make(chan struct{})
	wg.Add(1)
	go run(c, &wg)

	go func() {
		for res := range c.resutls {
			if res.error != nil {
				fmt.Fprintf(os.Stdout, "Error \t %s\n", res.error)
				continue
			}
			fmt.Fprintf(os.Stdout, "==>%s<==\n", res.fileName)
			fmt.Fprintln(os.Stdout, res.data)
		}
		doneCh <- struct{}{}
	}()

	wg.Wait()
	close(c.resutls)
	<-doneCh
}

func run(c Config, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(c.filesNames) == 0 {
		readDataFromReader(os.Stdin, c, "STDIN")
	} else {
		for _, file := range c.filesNames {
			wg.Add(1)
			go readData(file, c, wg)
		}
	}
}

func readData(fileName string, c Config, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := os.Open(fileName)
	if err != nil {
		c.resutls <- Result{
			fileName: fileName,
			error:    err,
		}
	}
	readDataFromReader(f, c, fileName)
}

func readDataFromReader(reader io.Reader, c Config, fileName string) {
	scanner := bufio.NewScanner(reader)
	//count lines has priority than count bytes
	maxCount := -1
	separator := ""

	if c.lines >= 0 {
		scanner.Split(bufio.ScanLines)
		maxCount = c.lines
		separator = "\n"
	} else if c.bytes >= 0 {
		scanner.Split(bufio.ScanBytes)
		maxCount = c.bytes
	} else {
		separator = "\n"
	}
	sb := strings.Builder{}
	if maxCount > -1 {
		for i := 0; scanner.Scan() && i < maxCount; i++ {
			sb.WriteString(scanner.Text() + separator)
		}
	} else {
		for scanner.Scan() {
			sb.WriteString(scanner.Text() + separator)
		}
	}
	c.resutls <- Result{
		fileName: fileName,
		error:    nil,
		data:     strings.TrimSuffix(sb.String(), separator),
	}
}
