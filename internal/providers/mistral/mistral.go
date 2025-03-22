package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"

	"github.com/appclacks/maizai/internal/otelspan"
	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	rag "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

type Config struct {
	APIKey string
}

type Client struct {
	config Config
	client *http.Client
}

func New(config Config) *Client {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
				return otelhttptrace.NewClientTrace(ctx)
			}),
		),
	}
	return &Client{
		config: config,
		client: httpClient,
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type queryPayload struct {
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature,omitempty"`
	Messages    []message `json:"messages"`
	MaxTokens   uint64    `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

type usage struct {
	PromptTokens     uint64 `json:"prompt_tokens"`
	CompletionTokens uint64 `json:"completion_tokens"`
	TotalTokens      uint64 `json:"total_tokens"`
}

type answerMessage struct {
	Prefix  bool   `json:"prefix"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type delta struct {
	Content string `json:"content"`
}

type choice struct {
	Index        uint64        `json:"index"`
	Message      answerMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
	Delta        delta         `json:"delta"`
}

type queryResponse struct {
	ID      string   `json:"id"`
	Usage   usage    `json:"usage"`
	Object  string   `json:"object"`
	Choices []choice `json:"choices"`
}

type embeddingData struct {
	Embedding []float32 `json:"embedding"`
}

type embeddingResponse struct {
	ID    string          `json:"id"`
	Usage usage           `json:"usage"`
	Data  []embeddingData `json:"data"`
}

type embeddingQuery struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

func eventData(body string) (*queryResponse, error) {
	after, found := strings.CutPrefix(body, "data: ")
	if !found {
		return nil, fmt.Errorf("invalid event data %s", body)
	}
	var result queryResponse
	err := json.Unmarshal([]byte(after), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (c *Client) Query(ctx context.Context, messages []shared.Message, options aggregates.QueryOptions) (*aggregates.Answer, error) {
	tracer := otel.Tracer("ai")
	ctx, span := tracer.Start(ctx, "Provider message")
	defer span.End()
	span.SetAttributes(semconv.GenAIRequestTemperature(options.Temperature))
	span.SetAttributes(semconv.GenAIRequestModel(options.Model))
	span.SetAttributes(semconv.GenAIRequestMaxTokens(int(options.MaxTokens)))
	span.SetAttributes(semconv.GenAISystemKey.String("mistral_ai"))
	payload := queryPayload{
		Model:       options.Model,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		Messages:    []message{},
	}
	if options.System != "" {
		message := message{
			Role:    "system",
			Content: options.System,
		}
		payload.Messages = append(payload.Messages, message)
	}
	for _, msg := range messages {
		message := message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		payload.Messages = append(payload.Messages, message)
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		otelspan.Error(span, err, "json error")
		return nil, err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.mistral.ai/v1/chat/completions",
		bytes.NewBuffer(jsonBytes))
	if err != nil {
		otelspan.Error(span, err, "fail to build mistral request")
		return nil, err

	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		otelspan.Error(span, err, "mistral api error")
		return nil, err
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		otelspan.Error(span, err, "http body error")
		return nil, err
	}
	if response.StatusCode >= 300 {
		err := fmt.Errorf("Mistral API returned an error: status %d\n%s", response.StatusCode, string(b))
		otelspan.Error(span, err, "mistral http error")
		return nil, err
	}
	var result queryResponse
	err = json.Unmarshal(b, &result)
	if err != nil {
		otelspan.Error(span, err, "json error")
		return nil, err
	}
	answer := aggregates.Answer{
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		Results:      []aggregates.Result{},
	}
	for _, choice := range result.Choices {
		answer.Results = append(answer.Results, aggregates.Result{
			Text: choice.Message.Content,
		})
	}
	span.SetAttributes(semconv.GenAIUsageInputTokens(int(result.Usage.PromptTokens)))
	span.SetAttributes(semconv.GenAIUsageOutputTokens(int(result.Usage.CompletionTokens)))
	span.SetStatus(codes.Ok, "success")
	return &answer, nil
}

func (c *Client) Embedding(ctx context.Context, query rag.EmbeddingQuery) (*rag.EmbeddingAnswer, error) {
	embeddingQuery := embeddingQuery{
		Model: query.Model,
		Input: []string{query.Input},
	}
	tracer := otel.Tracer("ai")
	ctx, span := tracer.Start(ctx, "Provider embedding")
	defer span.End()
	span.SetAttributes(semconv.GenAIRequestModel(query.Model))
	span.SetAttributes(semconv.GenAISystemKey.String("mistral_ai"))
	jsonBytes, err := json.Marshal(embeddingQuery)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.mistral.ai/v1/embeddings",
		bytes.NewBuffer(jsonBytes))
	if err != nil {
		otelspan.Error(span, err, "json error")
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		otelspan.Error(span, err, "mistral api error")
		return nil, err
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		otelspan.Error(span, err, "http body error")
		return nil, err
	}
	if response.StatusCode >= 300 {
		err := fmt.Errorf("Mistral API returned an error: status %d\n%s", response.StatusCode, string(b))
		otelspan.Error(span, err, "mistral http error")
		return nil, err
	}
	var result embeddingResponse
	err = json.Unmarshal(b, &result)
	if err != nil {
		otelspan.Error(span, err, "json error")
		return nil, err
	}
	answer := rag.EmbeddingAnswer{
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		Data:         []rag.Embedding{},
	}
	for _, data := range result.Data {
		answer.Data = append(answer.Data, rag.Embedding{
			Embedding: data.Embedding,
		})
	}
	span.SetAttributes(semconv.GenAIUsageInputTokens(int(result.Usage.PromptTokens)))
	span.SetStatus(codes.Ok, "success")
	return &answer, nil
}

func (c *Client) Stream(ctx context.Context, messages []shared.Message, options aggregates.QueryOptions) (<-chan aggregates.Event, error) {
	tracer := otel.Tracer("ai")
	ctx, span := tracer.Start(ctx, "Stream message")
	defer span.End()
	span.SetAttributes(semconv.GenAIRequestTemperature(options.Temperature))
	span.SetAttributes(semconv.GenAIRequestModel(options.Model))
	span.SetAttributes(semconv.GenAIRequestMaxTokens(int(options.MaxTokens)))
	span.SetAttributes(semconv.GenAISystemKey.String("mistral_ai"))
	payload := queryPayload{
		Model:       options.Model,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
		Messages:    []message{},
		Stream:      true,
	}
	if options.System != "" {
		message := message{
			Role:    "system",
			Content: options.System,
		}
		payload.Messages = append(payload.Messages, message)
	}
	for _, msg := range messages {
		message := message{
			Role:    msg.Role,
			Content: msg.Content,
		}
		payload.Messages = append(payload.Messages, message)
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		otelspan.Error(span, err, "json error")
		return nil, err
	}
	eventChan := make(chan aggregates.Event)
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.mistral.ai/v1/chat/completions",
		bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		otelspan.Error(span, err, "mistral api error")
		return nil, err
	}
	if response.StatusCode >= 300 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			otelspan.Error(span, err, "fail to read body")
			return nil, err
		}
		err = fmt.Errorf("Mistral API returned an error: status %d\n%s", response.StatusCode, string(body))
		otelspan.Error(span, err, "mistral http error")
		return nil, err
	}
	reader := bufio.NewReader(response.Body)

	go func() {
		ctx, bspan := tracer.Start(ctx, "Streaming background")
		defer bspan.End()
		defer response.Body.Close()
		finalMessage := ""
		var promptTokens, completionTokens uint64
		for {
			_, espan := tracer.Start(ctx, "Streaming event")
			line, err := reader.ReadBytes('\n')
			if err != nil {
				otelspan.Error(espan, err, "failed to read data")
				eventChan <- aggregates.Event{
					Error: err,
				}
				espan.End()
				break
			}
			lineStr := string(line)
			if lineStr == "" || lineStr == "\n" {
				continue
			}
			if strings.HasPrefix(lineStr, "data: [DONE]") {
				bspan.SetAttributes(semconv.GenAIUsageInputTokens(int(promptTokens)))
				bspan.SetAttributes(semconv.GenAIUsageOutputTokens(int(completionTokens)))
				results := []aggregates.Result{
					{
						Text: finalMessage,
					},
				}
				eventChan <- aggregates.Event{
					Answer: &aggregates.Answer{
						Results:      results,
						OutputTokens: completionTokens,
						InputTokens:  promptTokens,
					},
				}
				espan.SetStatus(codes.Ok, "success")
				espan.End()
				break
			}
			result, err := eventData(lineStr)
			if err != nil {
				otelspan.Error(espan, err, "failed to parse data")
				eventChan <- aggregates.Event{
					Error: err,
				}
				espan.End()
				break
			}
			for _, choice := range result.Choices {
				finalMessage = fmt.Sprintf("%s%s", finalMessage, choice.Delta.Content)
				eventChan <- aggregates.Event{
					Delta: choice.Delta.Content,
				}
			}
			if result.Usage.PromptTokens != 0 {
				promptTokens = result.Usage.PromptTokens
			}
			if result.Usage.CompletionTokens != 0 {
				completionTokens = result.Usage.CompletionTokens
			}
			bspan.SetStatus(codes.Ok, "success")
			espan.End()

		}

	}()
	span.SetStatus(codes.Ok, "success")
	return eventChan, nil
}
