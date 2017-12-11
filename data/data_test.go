package data_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/konjoot/find-go/data"
)

func TestDataSource_NewReadCloser(t *testing.T) {
	file, err := os.Open("./example.com")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	exampleCom, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, tc := range []struct {
		name    string
		srcType string
		target  string
		expOut  string
		expErr  error
	}{
		{
			name:    "URLSuccess",
			srcType: data.URL,
			target:  "http://example.com",
			expOut:  string(exampleCom),
		},
		{
			name:    "URLFail",
			srcType: data.URL,
			target:  "example.com",
			expOut:  "",
			expErr:  errors.New("Get example.com: unsupported protocol scheme \"\""),
		},
		{
			name:    "FileSuccess",
			srcType: data.File,
			target:  "./example.com",
			expOut:  string(exampleCom),
		},
		{
			name:    "FileFail",
			srcType: data.File,
			target:  "./example",
			expOut:  "",
			expErr:  errors.New("open ./example: no such file or directory"),
		},
		{
			name:    "WrongSrcType",
			srcType: "wrong",
			target:  "./example.com",
			expOut:  "",
			expErr:  errors.New("Unsupported data source type"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := data.NewSource(tc.srcType)

			rc, err := source.GetReadCloser(tc.target)
			t.Log("err =>", err)
			if fmt.Sprint(err) != fmt.Sprint(tc.expErr) {
				t.Error("Expected =>", tc.expErr)
				t.FailNow()
			}
			t.Log("rc =>", rc)
			if rc == nil {
				return
			}
			bts, err := ioutil.ReadAll(rc)
			t.Log("err =>", err)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			t.Log("output =>", string(bts))
			if string(bts) != tc.expOut {
				t.Error("Expected =>", tc.expOut)
			}
		})
	}
}
