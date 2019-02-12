package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultContentType = "application/json; charset=utf-8"
)

var (
	UnsupportMethodError = errors.New("UnsupportMethod")
)

// HttpClient 对http 客户端简单封装, 每个项目应该维护一个独立的client实例
type HttpClient struct {
	address string
	client  *http.Client
}

// Options optional args for HttpClient
type Options struct {
	Timeout             time.Duration
	IdleConnTimeout     time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	DialContext         func(ctx context.Context, network, addr string) (net.Conn, error)
}

type Option func(*Options)

func Timeout(timeout time.Duration) Option {
	return func(args *Options) {
		args.Timeout = timeout
	}
}
func MaxIdleConns(max int) Option {
	return func(args *Options) {
		args.MaxIdleConns = max
	}
}
func MaxIdleConnsPerHost(max int) Option {
	return func(args *Options) {
		args.MaxIdleConnsPerHost = max
	}
}

// NewClient build a HttpClient instance
func NewClient(address string, setters ...Option) *HttpClient {
	args := &Options{
		Timeout:             30 * time.Second,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 10,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   10 * time.Second,
		}).DialContext,
	}
	for _, setter := range setters {
		setter(args)
	}

	client := &http.Client{
		Timeout: args.Timeout,
		Transport: &http.Transport{
			DialContext:         args.DialContext,
			MaxIdleConns:        args.MaxIdleConns,
			MaxIdleConnsPerHost: args.MaxIdleConnsPerHost,
			IdleConnTimeout:     args.IdleConnTimeout,
		},
	}

	return &HttpClient{
		address: address,
		client:  client,
	}
}

func (c *HttpClient) SetAddr(addr string) *HttpClient {
	c.address = addr
	return c
}

// DoRequest
func (client *HttpClient) DoRequest(
	method, path string,
	query, header map[string]string,
	body []byte,
) ([]byte, error) {

	url := client.address + path

	var queryString []string
	for k, v := range query {
		queryString = append(queryString, fmt.Sprintf("%s=%s", k, v))
	}
	if len(queryString) > 0 {
		url += fmt.Sprintf("?%s", strings.Join(queryString, "&"))
	}

	log.Printf("ReqUrl: %s, Body: %s \n", url, string(body))

	var reader io.Reader
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
	case http.MethodPost, http.MethodPut:
		if body != nil {
			reader = bytes.NewReader(body)
		}
	default:
		return nil, UnsupportMethodError
	}

	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		request.Header.Set(k, v)
	}

	if reader != nil && request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", defaultContentType)
	}

	request.Close = true
	request.Header.Set("Connection", "close")
	response, err := client.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Get is http get method
func (client *HttpClient) Get(path string, query, header map[string]string) ([]byte, error) {
	return client.DoRequest(http.MethodGet, path, query, header, nil)
}

// Head is http head method
func (client *HttpClient) Head(path string, query, header map[string]string) ([]byte, error) {
	return client.DoRequest(http.MethodHead, path, query, header, nil)
}

// Delete is http delete method
func (client *HttpClient) Delete(path string, query, header map[string]string) ([]byte, error) {
	return client.DoRequest(http.MethodDelete, path, query, header, nil)
}

// Post is http post method
func (client *HttpClient) Post(path string, query, header map[string]string, body []byte) ([]byte, error) {
	return client.DoRequest(http.MethodPost, path, query, header, body)
}

// Put is http put method
func (client *HttpClient) Put(path string, query, header map[string]string, body []byte) ([]byte, error) {
	return client.DoRequest(http.MethodPut, path, query, header, body)
}
