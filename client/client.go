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
	Code    int    `json:"code"`
	Message string `json:"message"`
	KV      KV     `json:"kv"`
}

var (
	ErrAddrNotEmpty = errors.New("addr not empty")
	KeyNotEmpty     = errors.New("key not empty")
)

type Client struct {
	c         *http.Client
	masterUri string
	uris      []string
}

func NewClient(addr string, timeout time.Duration, idleConnTimeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) (*Client, error) {
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
	return &Client{c: c, masterUri: fmt.Sprintf("%s%s/kv", proto, addr)}, nil
}

func NewMSClient(master string, slaves []string, timeout time.Duration, idleConnTimeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) (*Client, error) {
	if master == "" {
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
	if !strings.HasPrefix(master, "http") {
		proto = "http://"
	}
	var uris []string
	for _, slave := range slaves {
		slaveUri := fmt.Sprintf("%s%s/kv", proto, slave)
		B.Register(slaveUri)
		uris = append(uris, slaveUri)
	}
	masterUri := fmt.Sprintf("%s%s/kv", proto, master)
	B.Register(masterUri)
	uris = append(uris, masterUri)
	return &Client{c: c, masterUri: masterUri, uris: uris}, nil
}

func New(addr string, timeout time.Duration) (*Client, error) {
	return NewClient(addr, timeout, 0, 0, 0)
}

func NewMS(master string, slaves []string, timeout time.Duration) (*Client, error) {
	return NewMSClient(master, slaves, timeout, 0, 0, 0)
}

func (c *Client) Put(key, value string) (*Result, error) {
	if key == "" {
		return nil, KeyNotEmpty
	}
	kv := &KV{K: key, V: value}
	body, err := json.Marshal(kv)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPut, c.masterUri, bytes.NewReader(body))
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

func (c *Client) Get(key string) (*Result, error) {
	return c.GetOrRealtime(key, false)
}

func (c *Client) GetRealtime(key string) (*Result, error) {
	return c.GetOrRealtime(key, true)
}

func (c *Client) GetOrRealtime(key string, realtime bool) (*Result, error) {
	if key == "" {
		return nil, KeyNotEmpty
	}
	uri := c.masterUri
	if realtime == false && c.uris != nil && len(c.uris) > 0 {
		uri = balancer.B.Choose()
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", uri, key), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.c.Do(request)
	if err != nil {
		return nil, err
	}
	if realtime == false {
		defer balancer.B.Increment()
	}
	defer response.Body.Close()
	return parseResponse(response)
}

func (c *Client) Delete(key string) (*Result, error) {
	if key == "" {
		return nil, KeyNotEmpty
	}
	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.masterUri, key), nil)
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
