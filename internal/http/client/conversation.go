package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type QueryOptions struct {
	Model       string         `json:"model"`
	System      string         `json:"system"`
	Temperature float64        `json:"temperature"`
	MaxTokens   uint64         `json:"max-tokens"`
	Provider    string         `json:"provider"`
	RagQuery    RagSearchQuery `json:"rag,omitempty"`
}

type ContextOptions struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Sources     ContextSources `json:"sources"`
}

type CreateConversationInput struct {
	QueryOptions      QueryOptions   `json:"query-options"`
	Prompt            string         `json:"prompt" required:"true"`
	ContextID         string         `json:"context-id,omitempty"`
	NewContextOptions ContextOptions `json:"new-context"`
	Stream            bool           `json:"stream"`
}

type Result struct {
	Text string `json:"text"`
}

type ConversationAnswer struct {
	Results      []Result `json:"result"`
	InputTokens  uint64   `json:"input-tokens"`
	OutputTokens uint64   `json:"output-tokens"`
	Context      string   `json:"context"`
}

type ConversationStreamEvent struct {
	Delta        string `json:"delta,omitempty"`
	Error        string `json:"error,omitempty"`
	InputTokens  uint64 `json:"input-tokens,omitempty"`
	OutputTokens uint64 `json:"output-tokens,omitempty"`
	Context      string `json:"context,omitempty"`
}

func (c *Client) CreateConversation(ctx context.Context, input CreateConversationInput) (*ConversationAnswer, error) {
	var result ConversationAnswer
	_, err := c.sendRequest(ctx, "/api/v1/conversation", http.MethodPost, input, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) StreamConversation(ctx context.Context, input CreateConversationInput) (<-chan ConversationStreamEvent, error) {
	eventChan := make(chan ConversationStreamEvent)
	var reqBody io.Reader
	j, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	reqBody = bytes.NewBuffer(j)
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/conversation", c.config.Endpoint),
		reqBody)
	if err != nil {
		return nil, err
	}
	request.Header.Add("content-type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 300 {
		b, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("the API returned an error: status %d\n%s", response.StatusCode, string(b))
	}
	reader := bufio.NewReader(response.Body)
	go func() {
		defer response.Body.Close()
		for {
			line, err := reader.ReadBytes('\n')
			lineStr := string(line)
			if err != nil {
				eventChan <- ConversationStreamEvent{
					Error: err.Error(),
				}
				close(eventChan)
				break
			}
			if lineStr == "" || lineStr == "\n" {
				continue
			}
			// todo check found
			after, _ := strings.CutPrefix(lineStr, "data: ")
			var event ConversationStreamEvent
			err = json.Unmarshal([]byte(after), &event)
			if err != nil {
				eventChan <- ConversationStreamEvent{
					Error: err.Error(),
				}
				close(eventChan)
				break
			}
			eventChan <- event
			if event.InputTokens != 0 {
				close(eventChan)
				break
			}
		}
	}()
	return eventChan, nil
}
