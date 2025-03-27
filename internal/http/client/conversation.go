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
	Model       string         `json:"model" required:"true" description:"The model to use"`
	System      string         `json:"system" description:"The system prompt"`
	Temperature float64        `json:"temperature" description:"The temperature parameter passed to the AI provider"`
	MaxTokens   uint64         `json:"max-tokens" required:"true" description:"The maximum number of tokens for the output"`
	Provider    string         `json:"provider" required:"true" description:"The AI provider to use"`
	RagQuery    RagSearchQuery `json:"rag,omitempty" description:"RAG query configuration"`
}

type ContextOptions struct {
	Name        string         `json:"name" required:"true" description:"The context name"`
	Description string         `json:"description" description:"The context description"`
	Sources     ContextSources `json:"sources" description:"Context sources to use for the new context"`
}

type CreateConversationInput struct {
	QueryOptions      QueryOptions   `json:"query-options" description:"The conversation query options"`
	Prompt            string         `json:"prompt" required:"true" description:"The prompt that will be passed to the AI provider"`
	ContextID         string         `json:"context-id,omitempty" description:"The ID of an existing context to use for this conversation"`
	NewContextOptions ContextOptions `json:"new-context" description:"Options to create a new context"`
	Stream            bool           `json:"stream" description:"Streaming mode using SSE"`
}

type Result struct {
	Text string `json:"text"`
}

type ConversationAnswer struct {
	Results      []Result `json:"result" description:"The result returned by the AI provider"`
	InputTokens  uint64   `json:"input-tokens" description:"The number of input tokens"`
	OutputTokens uint64   `json:"output-tokens" description:"The number of output tokens"`
	Context      string   `json:"context" description:"The ID of the context used for this conversation"`
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
