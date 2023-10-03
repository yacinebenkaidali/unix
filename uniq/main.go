package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Next Try to make the cli run on multiple files at once
// implement this using a concurrent safe map Mutexes

type Config struct {
	count     bool
	repeted   bool
	unique    bool
	stdSource bool
}

func main() {
	inputSource := flag.Bool("s", false, "read from stdin instead of from file")
	var count bool
	flag.BoolVar(&count, "c", false, "Count the occurrences of each of the lines")
	flag.BoolVar(&count, "count", false, "Count the occurrences of each of the lines")
	var repeated bool
	flag.BoolVar(&repeated, "d", false, "Print out only the repeated occurrences")
	flag.BoolVar(&repeated, "repeated", false, "Print out only the repeated occurrences")
	var uniq bool
	flag.BoolVar(&uniq, "u", false, "Print out only the unique occurrences")
	flag.BoolVar(&uniq, "uniq", false, "Print out only the unique occurrences")
	flag.Parse()

	var file string
	if len(flag.Args()) > 0 {
		file = flag.Args()[0]
	}

	cfg := Config{
		count:     count,
		stdSource: *inputSource,
		repeted:   repeated,
		unique:    uniq,
	}

	linesMap := map[string]int{}
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	var out io.Writer = getOutput(cfg.stdSource, file)
	if err := run(out, file, cfg, &linesMap); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out io.Writer, file string, cfg Config, linesMap *map[string]int) error {
	source, err := getInput(cfg.stdSource, file)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(source)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			count, exists := (*linesMap)[line]
			if exists {
				(*linesMap)[line] = count + 1
			} else {
				(*linesMap)[line] = 1
			}

		}
	}

	var output strings.Builder
	for line, count := range *linesMap {
		if cfg.repeted && count > 1 {
			if cfg.count {
				output.WriteString(fmt.Sprintf("%d %s\n", count, line))
			} else {
				output.WriteString(fmt.Sprintf("%s\n", line))
			}

		} else if cfg.unique && count == 1 {
			output.WriteString(fmt.Sprintf("%s\n", line))
		}
	}
	fmt.Fprint(out, output.String())
	return nil
}

func getOutput(stdinSrc bool, file string) io.Writer {
	if (stdinSrc) && file != "" {
		f, err := os.Create(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return f
	}
	return os.Stdout
}

func getInput(readFromStdin bool, fileName string) (io.Reader, error) {
	if readFromStdin {
		return os.Stdin, nil
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}
