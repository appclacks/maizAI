package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/appclacks/maizai/internal/otelspan"
	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

type Client struct {
	client *anthropic.Client
}

type Config struct {
	APIKey string
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
	client := anthropic.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithHTTPClient(httpClient),
	)
	return &Client{
		client: client,
	}
}

func (c *Client) Query(ctx context.Context, messages []shared.Message, options aggregates.QueryOptions) (*aggregates.Answer, error) {
	tracer := otel.Tracer("ai")
	ctx, span := tracer.Start(ctx, "Provider message")
	defer span.End()
	span.SetAttributes(semconv.GenAIRequestTemperature(options.Temperature))
	span.SetAttributes(semconv.GenAIRequestModel(options.Model))
	span.SetAttributes(semconv.GenAIRequestMaxTokens(int(options.MaxTokens)))
	span.SetAttributes(semconv.GenAISystemAnthropic)
	messagesParam := []anthropic.MessageParam{}
	for _, message := range messages {
		switch message.Role {
		case "user":
			messagesParam = append(messagesParam, anthropic.NewUserMessage(anthropic.NewTextBlock(message.Content)))
		case "assistant":
			messagesParam = append(messagesParam, anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Content)))
		default:
			err := fmt.Errorf("unknown role %s", message.Role)
			otelspan.Error(span, err, "unknown role")
			return nil, err
		}
	}
	messageParam := anthropic.MessageNewParams{
		Model:     anthropic.F(options.Model),
		MaxTokens: anthropic.F(int64(options.MaxTokens)),
		Messages:  anthropic.F(messagesParam),
	}
	if options.System != "" {
		messageParam.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(options.System),
		})
	}
	message, err := c.client.Messages.New(ctx, messageParam)
	if err != nil {
		otelspan.Error(span, err, "anthropic error")
		return nil, err
	}

	result := []aggregates.Result{}

	for _, m := range message.Content {
		result = append(result, aggregates.Result{
			Text: m.Text,
		})
	}
	answer := aggregates.Answer{
		Results:      result,
		InputTokens:  uint64(message.Usage.InputTokens),
		OutputTokens: uint64(message.Usage.OutputTokens),
	}
	span.SetAttributes(semconv.GenAIUsageInputTokens(int(message.Usage.InputTokens)))
	span.SetAttributes(semconv.GenAIUsageOutputTokens(int(message.Usage.OutputTokens)))
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
	span.SetAttributes(semconv.GenAISystemAnthropic)
	messagesParam := []anthropic.MessageParam{}
	for _, message := range messages {
		switch message.Role {
		case "user":
			messagesParam = append(messagesParam, anthropic.NewUserMessage(anthropic.NewTextBlock(message.Content)))
		case "assistant":
			messagesParam = append(messagesParam, anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Content)))
		default:
			err := fmt.Errorf("unknown role %s", message.Role)
			otelspan.Error(span, err, "unknown role")
			return nil, err
		}
	}
	messageParam := anthropic.MessageNewParams{
		Model:     anthropic.F(options.Model),
		MaxTokens: anthropic.F(int64(options.MaxTokens)),
		Messages:  anthropic.F(messagesParam),
	}
	if options.System != "" {
		messageParam.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(options.System),
		})
	}
	stream := c.client.Messages.NewStreaming(ctx, messageParam)
	eventChan := make(chan aggregates.Event)
	go func() {
		ctx, bspan := tracer.Start(ctx, "Streaming background")
		defer bspan.End()
		message := anthropic.Message{}
		for stream.Next() {
			_, espan := tracer.Start(ctx, "Streaming event")
			event := stream.Current()
			err := message.Accumulate(event)
			if err != nil {
				otelspan.Error(espan, err, "fail to accumulate message")
				eventChan <- aggregates.Event{
					Error: err,
				}
				break
			}
			switch delta := event.Delta.(type) {
			case anthropic.ContentBlockDeltaEventDelta:
				if delta.Text != "" {
					eventChan <- aggregates.Event{
						Delta: delta.Text,
					}
				}
			}
			espan.SetStatus(codes.Ok, "success")
			espan.End()
		}
		result := []aggregates.Result{}
		for _, m := range message.Content {
			result = append(result, aggregates.Result{
				Text: m.Text,
			})
		}
		eventChan <- aggregates.Event{
			Answer: &aggregates.Answer{
				Results:      result,
				OutputTokens: uint64(message.Usage.OutputTokens),
				InputTokens:  uint64(message.Usage.InputTokens),
			},
		}
		bspan.SetAttributes(semconv.GenAIUsageInputTokens(int(message.Usage.InputTokens)))
		bspan.SetAttributes(semconv.GenAIUsageOutputTokens(int(message.Usage.OutputTokens)))
		bspan.SetStatus(codes.Ok, "success")

		close(eventChan)
	}()
	span.SetStatus(codes.Ok, "success")
	return eventChan, nil
}
