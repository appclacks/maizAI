package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/appclacks/maizai/internal/tls"
)

type Client struct {
	http   *http.Client
	config Configuration
}

type Response struct {
	Messages []string `json:"messages"`
}

func New() (*Client, error) {
	client := &Client{
		http: &http.Client{},
	}
	config, err := Load()
	if err != nil {
		return nil, err
	}
	client.config = *config
	if config.Key != "" || config.Cert != "" || config.Cacert != "" || config.Insecure {
		tlsConfig, err := tls.GetTLSConfig(config.Key, config.Cert, config.Cacert, "", config.Insecure)
		if err != nil {
			return nil, err
		}
		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		client.http.Transport = transport
	}
	return client, nil
}

func (c *Client) sendRequest(ctx context.Context, url string, method string, body any, result any, queryParams map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		json, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(json)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.config.Endpoint, url),
		reqBody)
	if err != nil {
		return nil, err
	}
	if len(queryParams) != 0 {
		q := request.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}
	request.Header.Add("content-type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("the API returned an error: status %d\n%s", response.StatusCode, string(b))
	}
	if result != nil {
		err = json.Unmarshal(b, result)
		if err != nil {
			return nil, err
		}
	}
	return response, nil
}
