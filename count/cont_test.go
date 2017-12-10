package count

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestMatchCount(t *testing.T) {
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
			name:     "Golang100Times",
			str:      "Golang",
			pattern:  "Something about Golang golang go",
			count:    100,
			expCount: 100,
		},
		{
			name:     "Го10Times",
			str:      "Го",
			pattern:  "ГоПедагог",
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			r = bufio.NewReader(bytes.NewBufferString(
				strings.Repeat(tc.pattern, tc.count),
			))
			count, err = matchCount(tc.str, r)
			if err != io.EOF {
				t.Error(err)
			}
			t.Log("count =>", count)
			if count != tc.expCount {
				t.Error("Expected =>", tc.expCount)
			}
		})
	}
}
