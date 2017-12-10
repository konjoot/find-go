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

	subStringCounter := count.NewCounter(
		subString,
		poolSize,
		data.NewSource(srcType),
	)

	total, err := subStringCounter.CountMatches(ctx, os.Stdin, os.Stdout)
	fmt.Fprintf(os.Stdout, "Total: %d\n", total)
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
