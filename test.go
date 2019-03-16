package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"strings"
)

func main() {
	req := httptest.NewRequest("GET", "http://localhost", strings.NewReader("some_data"))

	bodyBuffer := bytes.NewBuffer(make([]byte, 0))
	teeReader := io.TeeReader(req.Body, bodyBuffer)
	req.Body = ioutil.NopCloser(teeReader)

	reqBody, err := ioutil.ReadAll(req.Body)
	fmt.Println("request read err", err)
	fmt.Println("request body", string(reqBody))

	reqBody, err = ioutil.ReadAll(bodyBuffer)
	fmt.Println("buffer read err", err)
	fmt.Println("buffer body", string(reqBody))
}
