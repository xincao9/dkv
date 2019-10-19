package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

type Result struct {
	Code    int    `json:code`
	Message string `json:message`
	KV      KV     `json:kv`
}

var (
	ErrAddrNotEmpty = errors.New("addr not empty")
)

type client struct {
	c   *http.Client
	uri string
}

func NewClient(addr string, timeout time.Duration, idleConnTimeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) (*client, error) {
	if addr == "" {
		return nil, ErrAddrNotEmpty
	}
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        maxIdleConns,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
			DisableCompression:  true,
		},
		Timeout: timeout,
	}
	proto := ""
	if !strings.HasPrefix(addr, "http") {
		proto = "http://"
	}
	return &client{c: c, uri: fmt.Sprintf("%s%s/kv", proto, addr)}, nil
}

func New(addr string, timeout time.Duration) (*client, error) {
	return NewClient(addr, timeout, 0, 0, 0)
}

func (c *client) Put(key, value string) (*Result, error) {
	if key == "" {
		return nil, errors.New("key not empty")
	}
	kv := &KV{K: key, V: value}
	body, err := json.Marshal(kv)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPut, c.uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("content-type", "application/json")
	response, err := c.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return parseResponse(response)
}

func (c *client) Get(key string) (*Result, error) {
	if key == "" {
		return nil, errors.New("key not empty")
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.uri, key), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return parseResponse(response)
}

func (c *client) Delete(key string) (*Result, error) {
	if key == "" {
		return nil, errors.New("key not empty")
	}
	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.uri, key), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return parseResponse(response)
}

func parseResponse(response *http.Response) (*Result, error) {
	r := &Result{}
	json.NewDecoder(response.Body).Decode(r)
	return r, nil
}
