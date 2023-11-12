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
	fileNames []string
	nbLines   bool
	blank     bool
	wg        *sync.WaitGroup
}

func main() {
	nbLines := flag.Bool("n", false, "Print out the line number")
	blank := flag.Bool("b", false, "Print out the line number")
	flag.Parse()

	c := Config{
		fileNames: flag.Args(),
		nbLines:   *nbLines,
		blank:     *blank,
	}
	c.wg = &sync.WaitGroup{}

	if err := run(&c); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

}

func run(c *Config) error {
	// go over the files and launch a goroutine for each of the files
	if len(c.fileNames) == 0 {
		//read from stdin
		res := readContent(os.Stdin, c.nbLines, c.blank)
		fmt.Print(res)
		return nil
	}
	//read from files
	for _, file := range c.fileNames {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		content := readContent(f, c.nbLines, c.blank)
		fmt.Print(content)
	}
	return nil
}

func readContent(in io.Reader, nbLines, blankLines bool) string {
	s := strings.Builder{}
	scanner := bufio.NewScanner(in)
	i := 1
	for scanner.Scan() {
		read := scanner.Text()
		if blankLines {
			if read != "" {
				s.WriteString(fmt.Sprintf("%d %s\n", i, read))
				i++
			} else {
				s.WriteString(fmt.Sprintf("%s\n", read))
			}
		} else if nbLines {
			s.WriteString(fmt.Sprintf("%d %s\n", i, read))
			i++
		} else {
			s.WriteString(fmt.Sprintf("%s\n", read))
		}
	}
	return s.String()
}
