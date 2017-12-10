package count

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/konjoot/find-go/data"
)

type Count struct {
	Count  int
	Target string
	Err    error
}

type Counter interface {
	CountMatches(context.Context, io.Reader, io.Writer) (int, error)
}

func NewCounter(subString string, poolSize int, source data.Source) Counter {
	return &counter{
		subString: subString,
		source:    source,
		pool:      make(chan struct{}, poolSize),
		countCh:   make(chan *Count, poolSize),
	}
}

type counter struct {
	subString string
	pool      chan struct{}
	source    data.Source
	countCh   chan *Count
}

func (c *counter) CountMatches(ctx context.Context, r io.Reader, w io.Writer) (int, error) {
	wg := c.countSubStrings(ctx, r)
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	bufw := bufio.NewWriter(w)
	var total int
	var count *Count
	for {
		select {
		case count = <-c.countCh:
			total += count.Count
			fmt.Fprintf(bufw, "Count for %s: %d\n", count.Target, count.Count)
			bufw.Flush()
			if count.Err != nil {
				return total, count.Err
			}
		case <-ctx.Done():
			return total, ctx.Err()
		case <-done:
			return total, nil
		}
	}
}

func (c *counter) countSubStrings(ctx context.Context, r io.Reader) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
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

			c.pool <- struct{}{}
			wg.Add(1)
			go func(target string) {
				defer func() {
					<-c.pool
					wg.Done()
				}()

				r, err := c.source.GetReader(target)
				if err != nil {
					return
				}
				defer r.Close()

				count, err := matchCount(c.subString, bufio.NewReader(r))
				c.countCh <- &Count{count, target, err}
			}(target)
		}
	}()
	return wg
}

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
