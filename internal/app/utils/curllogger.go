package utils

import (
	"bytes"
	"fmt"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/go-resty/resty/v2"
	"io"
	"net/http"
	"sort"
	"strings"
)

type Curl struct {
	slice []string
}

func (c *Curl) append(newSlice ...string) {
	c.slice = append(c.slice, newSlice...)
}

func (c *Curl) String() string {
	return strings.Join(c.slice, " ")
}

type nopCloser struct {
	io.Reader
}

func bashEscape(str string) string {
	return `'` + strings.ReplaceAll(str, `'`, `'\''`) + `'`
}

func (nopCloser) Close() error { return nil }

func logCurl(req *http.Request, excludeHeaders []string) (*Curl, error) {
	command := Curl{}

	command.append("curl")

	command.append("-X", bashEscape(req.Method))

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = nopCloser{bytes.NewBuffer(body)}
		bodyEscaped := bashEscape(string(body))
		command.append("-d", bodyEscaped)
	}

	var keys = make([]string, 0)

	for header := range req.Header {
		if !containsString(excludeHeaders, header) {
			keys = append(keys, header)
		}
	}

	sort.Strings(keys)

	for _, k := range keys {
		command.append("-H", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(req.Header[k], " "))))
	}

	command.append(bashEscape(req.URL.String()))

	return &command, nil
}

func containsString(sl []string, str string) bool {
	for _, s := range sl {
		if s == str {
			return true
		}
	}
	return false
}

func CurlLogger(_ *resty.Client, req *http.Request) error {
	logger := logdoc.GetLogger()
	command, _ := logCurl(req, []string{"Authorization"})
	logger.Debug(command)
	return nil
}
