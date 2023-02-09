package network

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

//goland:noinspection GoUnusedGlobalVariable,GoUnusedGlobalVariable,GoUnusedGlobalVariable
var (
	defaultHTTPClient = &http.Client{}
	HC10              = &http.Client{Timeout: time.Second * 10}
	HC15              = &http.Client{Timeout: time.Second * 15}
	HC30              = &http.Client{Timeout: time.Second * 30}
)

func ProcessGet(ctx context.Context, url string, headers map[string]string) ([]byte, int, error) {
	return unsafeHTTPCall(ctx, defaultHTTPClient, "GET", url, nil, headers)
}

func unsafeHTTPCall(ctx context.Context, client *http.Client, method string, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, 0, err
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}
