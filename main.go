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

// Application runner
func main() {
	// config
	var (
		srcType   string
		subString string
		duration  time.Duration
		poolSize  int
	)
	// flags
	flag.StringVar(&srcType, "type", data.URL, "data source type enum=[url|file]")
	flag.StringVar(&subString, "substring", "Go", "substring for counting")
	flag.DurationVar(&duration, "duration", time.Minute, "execution time limit")
	flag.IntVar(&poolSize, "pool-size", runtime.GOMAXPROCS(0), "limits goroutines max count")
	flag.Parse()

	if poolSize < 1 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", "pool-size should be greater than 0")
		os.Exit(1)
	}

	// application wide context initialization
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// termination listener
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		// stops application when
		select {
		case <-stop:
			// termination signal appeared
		case <-ctx.Done():
			// context cancelled
		}
		cancel()
	}()

	// counter initialization
	counter := count.NewSubStringCounter(
		subString,
		poolSize,
		data.NewSource(srcType),
	)

	// counts total and prints results
	total, err := CountTotal(ctx, counter, os.Stdin, os.Stdout)
	fmt.Fprintf(os.Stdout, "Total: %d\n", total)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// CountTotal uses counter for substrings counting
// reads targers from io.Reader,
// writes counts for every target into the io.Writer
// returns total count and last counting error
func CountTotal(ctx context.Context, counter count.Counter, r io.Reader, w io.Writer) (int, error) {
	wg := counter.Count(ctx, r)
	// done is a cancellation channel
	done := make(chan struct{})
	// waits for all routines in counter.Count to finish
	// and closes done channel to finish the app
	go func() { wg.Wait(); close(done) }()

	var (
		err   error
		total int
		count *count.Count
	)
	for {
		select {
		case count = <-counter.CountCh():
			// new count appeared
			if count.Err != nil {
				err = count.Err // remember only last error
			}
			total += count.Count
			fmt.Fprintf(w, "Count for %s: %d\n", count.Target, count.Count)
		case <-ctx.Done():
			// context cancelled
			return total, ctx.Err()
		case <-done:
			// all work is done
			return total, err
		}
	}
}
