package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync/atomic"
)

type HttpClient struct {
	idCounter uint32
	rootCtx   context.Context
	client    *http.Client
	headers   http.Header
	url       string
}

func DialHTTP(endpoint string) (Client, error) {
	return DialHTTPWithClient(endpoint, new(http.Client))
}

func DialHTTPWithClient(endpoint string, client *http.Client) (*HttpClient, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	headers := make(http.Header, 2)
	headers.Set("accept", contentType)
	headers.Set("content-type", contentType)

	return &HttpClient{
		client:  client,
		headers: headers,
		url:     endpoint,
		rootCtx: context.Background(),
	}, nil
}

func (h *HttpClient) Close() {
	return
}

func (h *HttpClient) Call(result interface{}, url string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(h.rootCtx, DefaultReadTimeout+DefaultWriteTimeout)
	defer cancel()
	return h.CallContext(ctx, result, url, args...)
}

func (h *HttpClient) buildMsg(args ...interface{}) (json.RawMessage, error) {
	id := atomic.AddUint32(&h.idCounter, 1)
	msg := &JsonRpcMessageItem{Version: vsn, ID: id}
	if args != nil {
		var err error
		if msg.Params, err = json.Marshal(args); err != nil {
			return nil, err
		}
	}
	ret := make(JsonRpcMessage, 1)
	ret[0] = msg
	body, err := json.Marshal(ret)
	if err != nil {
		return nil, err
	}
	return body, nil
}
func (h *HttpClient) buildRequest(ctx context.Context, path string, body []byte) (*http.Request, error) {

	req, err := http.NewRequestWithContext(ctx,
		"POST",
		h.url+path,
		ioutil.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}

	req.ContentLength = int64(len(body))
	req.Header = h.headers.Clone()
	return req, nil
}

func (h *HttpClient) CallContext(ctx context.Context, result interface{}, path string, args ...interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}
	body, err := h.buildMsg(args)
	if err != nil {
		return err
	}

	req, err := h.buildRequest(ctx, path, body)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(out, &result)

}