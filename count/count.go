package count

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"

	"github.com/konjoot/find-go/data"
)

// Count represents result for counting a substring into the source
type Count struct {
	Count  int
	Target string
	Err    error
}

// Counter is an interface for every match counter
type Counter interface {
	// Count counts implementation related matches, e.g. substrings
	Count(ctx context.Context, r io.Reader) *sync.WaitGroup
	// CountCh returns chan with counts
	CountCh() <-chan *Count
}

// NewSubStringCounter constructs substring Counter
func NewSubStringCounter(subString string, poolSize int, source data.Source) Counter {
	return &subStringCounter{
		subString: subString,
		source:    source,
		pool:      make(chan struct{}, poolSize),
		countCh:   make(chan *Count, poolSize),
	}
}

// subStringCounter is a substring Counter
type subStringCounter struct {
	subString string
	pool      chan struct{}
	source    data.Source
	countCh   chan *Count
}

// Count reads targers from reader and sends them into separate routines
// where they converts into the counter results.
func (sc *subStringCounter) Count(ctx context.Context, r io.Reader) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	go sc.count(wg)(ctx, r)

	return wg
}

// CountCh returns chan with counts
func (sc *subStringCounter) CountCh() <-chan *Count {
	return sc.countCh
}

// count is a constructor for the target's emitter
func (sc *subStringCounter) count(wg waitGrouper) func(context.Context, io.Reader) {
	// incremet wait group counter
	wg.Add(1)
	// target's emitter, which reads lines(targets) from
	// io.Reader and schedule substring counting for them
	return func(ctx context.Context, r io.Reader) {
		defer wg.Done() // on exit decrement wait group counter

		var (
			target string
			err    error
			bufr   = bufio.NewReader(r)
		)
		for err == nil {
			target, err = bufr.ReadString('\n')
			target = strings.TrimSuffix(target, "\n")
			if target == "" {
				return
			}
			if err != nil && err != io.EOF {
				sc.countCh <- &Count{0, target, err}
				return
			}

			go sc.readAndCount(wg)(target)
		}
	}
}

// readAndCount is a constructor of the substrings counter
func (sc *subStringCounter) readAndCount(wg waitGrouper) func(string) {
	// wait for a place in the pool
	// the number of routines is limited by the size of the pool
	sc.pool <- struct{}{}
	// increment wait group counter
	wg.Add(1)

	// substring counter,
	// gets target as a string,
	// gets it's data from data.Source,
	// counts substrings ans writes them into sc.countCh
	return func(target string) {
		// on exit
		defer func() {
			// release place in the pool
			<-sc.pool
			// decrement wait group counter
			wg.Done()
		}()

		r, err := sc.source.GetReadCloser(target)
		if err != nil {
			sc.countCh <- &Count{0, target, err}
			return
		}
		defer r.Close()

		count, err := countSubStrings(sc.subString, bufio.NewReader(r))
		sc.countCh <- &Count{count, target, err}
	}
}

// countSubStrings counts substring matches from io.Reader.
// Not an efficient solution, but robust enough.
func countSubStrings(str string, r io.RuneReader) (int, error) {
	var (
		count int
		rn    rune
		buf   = make([]rune, len([]rune(str)))
		err   error
	)
	if len([]rune(str)) == 0 {
		return 0, nil
	}
	for err == nil {
		rn, _, err = r.ReadRune()
		buf = append(buf[1:], rn)
		if str == string(buf) {
			count++
		}
	}
	if err != nil && err != io.EOF {
		return count, err
	}

	return count, nil
}

// waitGrouper is an interface for sync.WaitGroup
type waitGrouper interface {
	Add(int)
	Done()
}
