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
	// CountSubStrings counts implementation related matches,
	// e.g. substrings and connects reader(input) and writer(output)
	CountSubStrings(ctx context.Context, r io.Reader) *sync.WaitGroup
	// CountCh returns chan with counts
	CountCh() <-chan *Count
}

// NewCounter constructs substring Counter
func NewCounter(subString string, poolSize int, source data.Source) Counter {
	return &counter{
		subString: subString,
		source:    source,
		pool:      make(chan struct{}, poolSize),
		countCh:   make(chan *Count, poolSize),
	}
}

// counter is a substring Counter
type counter struct {
	subString string
	pool      chan struct{}
	source    data.Source
	countCh   chan *Count
}

// CountSubStrings reads targers from reader and sends them into separate routines
// where they converts into the counter results. All parallelism is here.
func (c *counter) CountSubStrings(ctx context.Context, r io.Reader) *sync.WaitGroup {
	wg := new(sync.WaitGroup)

	go c.startCounting(wg)(ctx, r)

	return wg
}

// CountCh returns chan with counts
func (c *counter) CountCh() <-chan *Count {
	if c.countCh != nil {
		return c.countCh
	}
	c.countCh = make(chan *Count, len(c.pool))
	return c.countCh
}

func (c *counter) startCounting(wg *sync.WaitGroup) func(context.Context, io.Reader) {
	wg.Add(1)
	return func(ctx context.Context, r io.Reader) {
		defer wg.Done()

		bufr := bufio.NewReader(r)
		var (
			target string
			err    error
		)
		for err == nil {
			target, err = bufr.ReadString('\n')
			target = strings.TrimSuffix(target, "\n")
			if err == io.EOF {
				return
			}
			if err != nil {
				c.countCh <- &Count{0, target, err}
				return
			}

			go c.readAndCount(wg)(target)
		}
	}
}

func (c *counter) readAndCount(wg *sync.WaitGroup) func(string) {
	c.pool <- struct{}{}
	wg.Add(1)

	return func(target string) {
		defer func() {
			<-c.pool
			wg.Done()
		}()

		r, err := c.source.GetReader(target)
		if err != nil {
			c.countCh <- &Count{0, target, err}
			return
		}
		defer r.Close()

		count, err := matchCount(c.subString, bufio.NewReader(r))
		c.countCh <- &Count{count, target, err}
	}
}

// matchCount counts substring matches from io.Reader
func matchCount(str string, r io.RuneReader) (int, error) {
	var (
		count int
		buf   = make([]rune, len([]rune(str)))
		rn    rune
		err   error
	)
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
