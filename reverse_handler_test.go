package mps

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReverseHandler(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	reverseHandler := NewReverseHandler()
	proxySrv := httptest.NewServer(reverseHandler)
	defer proxySrv.Close()

	resp, err := HttpGet(srv.URL, func(r *http.Request) (*url.URL, error) {
		return url.Parse(proxySrv.URL)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodySize := len(body)
	contentLength, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	asserts := assert.New(t)
	asserts.Equal(resp.StatusCode, 200, "statusCode should be equal 200")
	asserts.Equal(bodySize, contentLength, "Content-Length should be equal "+strconv.Itoa(bodySize))
	asserts.Equal(int64(bodySize), resp.ContentLength)
}
