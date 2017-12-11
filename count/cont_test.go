package count

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
)

func TestCounter_count(t *testing.T) {

	for _, tc := range []struct {
		name      string
		subString string
		poolSize  int
		ctx       context.Context
		reader    io.Reader
		source    string
		sourceErr error
		expCounts map[string]*Count
	}{
		{
			name:      "FindGoSuccess",
			subString: "Go",
			poolSize:  2,
			ctx:       context.Background(),
			reader:    bytes.NewBufferString("one\ntwo\nthree\nfour\nfive"),
			source:    "GogoGo",
			sourceErr: nil,
			expCounts: map[string]*Count{
				"one":   &Count{2, "one", nil},
				"two":   &Count{2, "two", nil},
				"three": &Count{2, "three", nil},
				"four":  &Count{2, "four", nil},
				"five":  &Count{2, "five", nil},
			},
		},
		{
			name:      "FindGoDataSourceErr",
			subString: "Go",
			poolSize:  4,
			ctx:       context.Background(),
			reader:    bytes.NewBufferString("one\ntwo\n\nthree"),
			source:    "GoGo",
			sourceErr: errors.New("test"),
			expCounts: map[string]*Count{
				"one":   &Count{0, "one", errors.New("test")},
				"two":   &Count{0, "two", errors.New("test")},
				"three": &Count{0, "three", errors.New("test")},
			},
		},
	} {

		t.Run(tc.name, func(t *testing.T) {

			dataSource := &mockDataSource{tc.source, tc.sourceErr}
			counter := NewSubStringCounter(
				tc.subString,
				tc.poolSize,
				dataSource,
			).(*subStringCounter)

			wg := &sync.WaitGroup{}

			go counter.count(wg)(tc.ctx, tc.reader)

			done := make(chan struct{})
			go func() { wg.Wait(); close(done) }()

			for {
				select {
				case count := <-counter.CountCh():
					t.Log("count.Count =>", count.Count)
					if count.Count != tc.expCounts[count.Target].Count {
						t.Error("Expected =>", tc.expCounts[count.Target].Count)
					}
					t.Log("count.Err =>", count.Err)
					if fmt.Sprint(count.Err) != fmt.Sprint(tc.expCounts[count.Target].Err) {
						t.Error("Expected =>", tc.expCounts[count.Target].Err)
					}
				case <-done:
					return
				}
			}

		})
	}
}

// To ensure that pool perform as expected
func TestCounter_readAndCount(t *testing.T) {
	for _, tc := range []struct {
		name          string
		poolSize      int
		routinesCount int
		expWgMax      int
		expErr        error
	}{
		{
			name:          "23RoutinesBy4",
			poolSize:      4,
			routinesCount: 23,
			expWgMax:      4,
			expErr:        nil,
		},
		{
			name:          "23RoutinesBy90",
			poolSize:      90,
			routinesCount: 23,
			expWgMax:      23,
			expErr:        nil,
		},
		{
			name:          "13RoutinesBy1",
			poolSize:      1,
			routinesCount: 13,
			expWgMax:      1,
			expErr:        nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dataSource := &mockDataSource{}
			counter := NewSubStringCounter(
				"",
				tc.poolSize,
				dataSource,
			).(*subStringCounter)

			wg := &mockWaitGroup{WaitGroup: sync.WaitGroup{}}

			go func() {
				for range counter.CountCh() {
				}
			}()

			for i := 0; i < tc.routinesCount; i++ {
				go counter.readAndCount(wg)("target")
			}

			wg.WaitGroup.Wait()

			t.Log("wg.max =>", wg.max)
			if wg.max > tc.expWgMax {
				t.Error("Expected =>", tc.expWgMax)
			}
		})
	}
}

func Test_countSubStrings(t *testing.T) {
	var (
		r     *bufio.Reader
		count int
		err   error
	)

	for _, tc := range []struct {
		name     string
		str      string
		count    int
		pattern  string
		expCount int
	}{
		{
			name:     "Go",
			str:      "Golang",
			pattern:  "Something about Golang golang go",
			count:    100,
			expCount: 100,
		},
		{
			name:     "Го10Times",
			str:      "Го",
			pattern:  "Голанг",
			count:    10,
			expCount: 10,
		},
		{
			name:     "NoMatches",
			str:      "test",
			pattern:  "hidden tes\t",
			count:    1,
			expCount: 0,
		},
		{
			name:     "EmptyString",
			str:      "",
			pattern:  "hidden tes\t",
			count:    10,
			expCount: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r = bufio.NewReader(bytes.NewBufferString(
				strings.Repeat(tc.pattern, tc.count),
			))
			count, err = countSubStrings(tc.str, r)
			if err != nil {
				t.Error(err)
			}
			t.Log("count =>", count)
			if count != tc.expCount {
				t.Error("Expected =>", tc.expCount)
			}
		})
	}
}

type mockDataSource struct {
	source    string
	sourceErr error
}

func (ds *mockDataSource) GetReadCloser(target string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBufferString(ds.source)), ds.sourceErr
}

type mockWaitGroup struct {
	sync.WaitGroup
	sync.Mutex
	counter int
	max     int
}

func (wg *mockWaitGroup) Add(delta int) {
	wg.Lock()
	wg.counter += delta
	if wg.max < wg.counter {
		wg.max = wg.counter
	}
	wg.Unlock()
	wg.WaitGroup.Add(delta)
}

func (wg *mockWaitGroup) Done() {
	wg.Lock()
	wg.counter--
	wg.Unlock()
	wg.WaitGroup.Done()
}
