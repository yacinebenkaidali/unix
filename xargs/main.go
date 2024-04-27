package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Config struct {
	n     int
	p     int
	wg    *sync.WaitGroup
	errCh chan error
}

type intervals struct {
	bounderies [][2]int
}

func main() {
	n := flag.Int("n", 5000, "Set the maximum number of arguments taken from standard input for each invocation of utility")
	p := flag.Int("P", 1, "Parallel mode: run at most maxprocs invocations of utility at once")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("need to pass at least a command")
	}
	c := Config{
		n:     *n,
		p:     *p,
		wg:    &sync.WaitGroup{},
		errCh: make(chan error),
	}

	if c.n < 1 {
		log.Fatal("n has to be greater or equal 1")
	}

	go func() {
		err := <-c.errCh
		log.Println(err)
	}()

	now := time.Now()
	defer func() {
		fmt.Printf("%+v\n", time.Since(now))
	}()

	if err := run(&c); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	c.wg.Wait()
	close(c.errCh)
}

func run(c *Config) error {
	mainCmd := flag.Arg(0)
	cmdArgs := flag.Args()[1:]

	path, err := exec.LookPath(mainCmd)
	if err != nil {
		return err
	}
	_ = path
	args := readArgs()
	allArgs := []string{}
	allArgs = append(allArgs, cmdArgs...)
	allArgs = append(allArgs, args...)

	nbInvocations := math.Ceil(float64(len(args)) / float64(c.n))
	intervals := intervals{
		bounderies: make([][2]int, int(nbInvocations)),
	}
	from := 0
	to := min(from+c.n, len(allArgs))
	intervals.bounderies[0] = [2]int{from, to}

	for i := 1; i < int(nbInvocations); i++ {
		from = to
		to = min(from+c.n, len(allArgs))
		intervals.bounderies[i] = [2]int{from, to}
	}
	//spawn `P` goroutines
	// fmt.Println(intervals.bounderies)
	processPortion := len(intervals.bounderies)
	from = 0
	to = 0

	for i := 0; i < c.p; i++ {
		currentPortion := math.Ceil(float64(processPortion) / float64(c.p-i))
		from = to
		to += int(currentPortion)

		//launch a goroutine
		c.wg.Add(1)
		go func(goid, from, to int) {
			defer c.wg.Done()
			for i := 0; i < len(intervals.bounderies[from:to]); i++ {
				interval := intervals.bounderies[from:to][i]
				args := allArgs[interval[0]:interval[1]]
				cmd := exec.Command(path, args...)
				if len(args) > 0 {
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						c.errCh <- err
					}
				}
				// fmt.Println(cmd.String(), interval, "go", goid)
			}
		}(i, from, to)
		processPortion -= int(math.Ceil(float64(processPortion) / float64(c.p-i)))
	}

	return nil
}

func readArgs() []string {
	scanner := bufio.NewScanner(os.Stdin)
	args := []string{}
	for scanner.Scan() {
		args = append(args, scanner.Text())
	}

	return args
}
