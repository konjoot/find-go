# Find-go
Substrings counter. The app gets strings from StdIn, reads data from sources, which they are represents and counts target substrings into those data.

# Installation

`go get -u github.com/konjoot/find-go/...`

# Usage

Use `-h` to get help about flags

```
Usage of find-go:
  -duration duration
    	execution time limit (default 1m0s)
  -pool-size int
    	limits goroutines max count (default 4)
  -substring string
    	substring for counting (default "Go")
  -type string
    	data source type enum=[url|file] (default "url")
```

Count substrings in data available from URLs:

`echo -e 'https://golang.org\nhttps://golang.org\nhttps://golang.org/pkg/strings/' | find-go`

Count substrings in files:

`echo -e '/etc/passwd\n/etc/hosts' | find-go -type=file`

# Testing

Ensure you are in the project's root: `cd $GOPATH/src/github.com/konjoot/find-go/`

To run tests `go test ./...`
