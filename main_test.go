package find_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	find "github.com/konjoot/find-go"
	"github.com/konjoot/find-go/count"
	"github.com/konjoot/find-go/data"
)

func TestCountTotal(t *testing.T) {
	for _, tc := range []struct {
		name      string
		subString string
		poolSize  int
		srcType   string
		input     []string
		output    []string
		expErr    error
		expTotal  int
	}{
		{
			name:      "URLsSuccess",
			subString: "Go",
			poolSize:  4,
			srcType:   data.URL,
			input: []string{
				"https://golang.org",
				"https://golang.org",
				"https://golang.org",
			},
			output: []string{
				"Count for https://golang.org: 9",
				"Count for https://golang.org: 9",
				"Count for https://golang.org: 9",
			},
			expTotal: 27,
			expErr:   nil,
		},
		{
			name:      "URLsSuccessUTF8",
			subString: "Го",
			poolSize:  4,
			srcType:   data.URL,
			input: []string{
				"https://golang.org",
				"https://golang.org",
				"https://golang.org",
			},
			output: []string{
				"Count for https://golang.org: 0",
				"Count for https://golang.org: 0",
				"Count for https://golang.org: 0",
			},
			expTotal: 0,
			expErr:   nil,
		},
		{
			name:      "URLsFail",
			subString: "Go",
			poolSize:  4,
			srcType:   data.URL,
			input: []string{
				"golang.org",
				"https://golang.org",
				"https://golang.org",
			},
			output: []string{
				"Count for golang.org: 0",
				"Count for https://golang.org: 9",
				"Count for https://golang.org: 9",
			},
			expTotal: 18,
			expErr:   errors.New("Get golang.org: unsupported protocol scheme \"\""),
		},
		{
			name:      "FilesSuccess",
			subString: "Go",
			poolSize:  4,
			srcType:   data.File,
			input: []string{
				"/etc/passwd",
				"/etc/hosts",
			},
			output: []string{
				"Count for /etc/passwd: 0",
				"Count for /etc/hosts: 0",
			},
			expTotal: 0,
			expErr:   nil,
		},
		{
			name:      "FilesSuccessUTF8",
			subString: "Го",
			poolSize:  4,
			srcType:   data.File,
			input: []string{
				"/etc/passwd",
				"/etc/hosts",
			},
			output: []string{
				"Count for /etc/passwd: 0",
				"Count for /etc/hosts: 0",
			},
			expTotal: 0,
			expErr:   nil,
		},
		{
			name:      "FilesFail",
			subString: "Go",
			poolSize:  4,
			srcType:   data.File,
			input: []string{
				"/some-place",
				"/etc/hosts",
			},
			output: []string{
				"Count for /some-place: 0",
				"Count for /etc/hosts: 0",
			},
			expTotal: 0,
			expErr:   errors.New("open /some-place: no such file or directory"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewBufferString(strings.Join(tc.input, "\n"))
			w := bytes.NewBuffer([]byte{})

			counter := count.NewSubStringCounter(
				tc.subString,
				tc.poolSize,
				data.NewSource(tc.srcType),
			)
			total, err := find.CountTotal(context.Background(), counter, r, w)
			t.Log("err =>", err)
			if fmt.Sprint(err) != fmt.Sprint(tc.expErr) {
				t.Error("Expected =>", tc.expErr)
			}
			t.Log("total =>", total)
			if total != tc.expTotal {
				t.Error("Expected =>", tc.expTotal)
			}

			sort.Strings(tc.output)
			outputStrings := strings.Split(strings.TrimSuffix(w.String(), "\n"), "\n")
			sort.Strings(outputStrings)
			output := strings.Join(outputStrings, "\n")
			t.Log("output =>", output)
			if output != strings.Join(tc.output, "\n") {
				t.Error("Expected =>", strings.Join(tc.output, "\n"))
			}
		})
	}
}
