package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	client client.MCPClient
}

func NewSSEClient(ctx context.Context, url string) (*Client, error) {
	c, err := client.NewSSEMCPClient(url)
	if err != nil {
		return nil, err
	}
	err = c.Start(ctx)
	if err != nil {
		return nil, err
	}
	initRequest := mcp.InitializeRequest{}
	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func NewStdioClient(ctx context.Context, command string, env []string, args ...string) (*Client, error) {
	c, err := client.NewStdioMCPClient(command, env, args...)
	if err != nil {
		return nil, err
	}
	initRequest := mcp.InitializeRequest{}
	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func (c *Client) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	toolsRequest := mcp.ListToolsRequest{}
	result, err := c.client.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (c *Client) CallTool(ctx context.Context, name string, params map[string]any) ([]mcp.Content, error) {
	toolRequest := mcp.CallToolRequest{}
	toolRequest.Params.Name = name
	toolRequest.Params.Arguments = params

	result, err := c.client.CallTool(ctx, toolRequest)
	if err != nil {
		return nil, err
	}
	if result.IsError {
		return nil, fmt.Errorf("fail to call tool %s", name)
	}
	return result.Content, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx)
}
