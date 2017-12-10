package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/konjoot/find-go/count"
	"github.com/konjoot/find-go/data"
)

func main() {
	var (
		srcType   string
		subString string
		duration  time.Duration
		poolSize  int
	)
	flag.StringVar(&srcType, "type", data.URL, "data source type enum=[url|file]")
	flag.StringVar(&subString, "substring", "Go", "substring for counting")
	flag.DurationVar(&duration, "duration", time.Minute, "execution time limit")
	flag.IntVar(&poolSize, "pool-size", runtime.GOMAXPROCS(0), "limits goroutines max count")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		select {
		case <-stop:
		case <-ctx.Done():
		}
		cancel()
	}()

	counter := count.NewCounter(
		subString,
		poolSize,
		data.NewSource(srcType),
	)

	total, err := countTotal(ctx, counter, os.Stdin, os.Stdout)
	fmt.Fprintf(os.Stdout, "Total: %d\n", total)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func countTotal(ctx context.Context, counter count.Counter, r io.Reader, w io.Writer) (int, error) {
	wg := counter.CountSubStrings(ctx, r)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	var err error
	var total int
	var count *count.Count
	for {
		select {
		case count = <-counter.CountCh():
			err = count.Err // remember only last error
			total += count.Count
			fmt.Fprintf(w, "Count for %s: %d\n", count.Target, count.Count)
		case <-ctx.Done():
			return total, ctx.Err()
		case <-done:
			return total, err
		}
	}
}
