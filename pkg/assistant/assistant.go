package assistant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	ragdata "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
)

var ragPlaceholder = "{maizai_rag_data}"

type Provider interface {
	Query(ctx context.Context, messages []shared.Message, options aggregates.QueryOptions) (*aggregates.Answer, error)
}

type ContextManager interface {
	CreateOrGetContext(ctx context.Context, contextID string, options shared.ContextOptions) (*shared.Context, error)
	GetContext(ctx context.Context, id string) (*shared.Context, error)
	AddMessagesToContext(ctx context.Context, id string, messages []shared.Message) error
}

type Rag interface {
	Match(ctx context.Context, query ragdata.SearchQuery) ([]ragdata.DocumentChunk, error)
}

type Assistant struct {
	rag        Rag
	ctxManager ContextManager
	providers  map[string]Provider
}

func New(clients map[string]Provider, ctxManager ContextManager, rag Rag) *Assistant {
	return &Assistant{
		rag:        rag,
		ctxManager: ctxManager,
		providers:  clients,
	}
}

func (a *Assistant) Message(ctx context.Context, messages []shared.Message, options aggregates.QueryOptions) (*aggregates.Answer, error) {
	client, ok := a.providers[options.Provider]
	if !ok {
		return nil, fmt.Errorf("AI client %s not found", options.Provider)
	}
	answer, err := client.Query(ctx, messages, options)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (a *Assistant) enrichRecursively(ctx context.Context, context *shared.Context, messages []shared.Message) ([]shared.Message, error) {
	result := []shared.Message{}
	if context.Sources.Contexts != nil {
		for _, source := range context.Sources.Contexts {
			sourceContext, err := a.ctxManager.GetContext(ctx, source)
			if err != nil {
				return nil, err
			}
			msg, err := a.enrichRecursively(ctx, sourceContext, messages)
			if err != nil {
				return nil, err
			}
			result = append(result, msg...)
		}
	}

	result = append(result, context.Messages...)
	return result, nil
}

func (a *Assistant) Enrich(ctx context.Context, context *shared.Context, messages []shared.Message) ([]shared.Message, error) {
	result, err := a.enrichRecursively(ctx, context, messages)
	if err != nil {
		return nil, err
	}
	result = append(result, messages...)
	return result, nil
}

func (a *Assistant) UpdateContext(ctx context.Context, context string, messages []shared.Message, results []aggregates.Result) error {
	update := []shared.Message{}
	update = append(update, messages...)
	for _, result := range results {
		id, err := uuid.NewV6()
		if err != nil {
			return err
		}
		update = append(update, shared.Message{
			ID:        id.String(),
			CreatedAt: time.Now().UTC(),
			Role:      shared.AssistantRole,
			Content:   result.Text,
		})
	}
	return a.ctxManager.AddMessagesToContext(ctx, context, update)
}

func (a *Assistant) EnrichWithRag(ctx context.Context, messages []shared.Message, ragQuery ragdata.SearchQuery) ([]shared.Message, error) {

	chunks, err := a.rag.Match(ctx, ragQuery)
	fragments := []string{}
	for _, chunk := range chunks {
		fragments = append(fragments, chunk.Fragment)
	}
	if err != nil {
		return nil, err
	}
	ragData := strings.Join(fragments, "\n")
	result := []shared.Message{}
	for _, message := range messages {
		message.Content = strings.ReplaceAll(message.Content, ragPlaceholder, ragData)
		result = append(result, message)
	}
	return result, nil

}

func (a *Assistant) Pipeline(
	ctx context.Context,
	options aggregates.QueryOptions,
	contextOptions shared.ContextOptions,
	contextID string,
	messages []shared.Message) (*aggregates.Answer, error) {
	context, err := a.ctxManager.CreateOrGetContext(ctx, contextID, contextOptions)
	if err != nil {
		return nil, err
	}
	for _, m := range messages {
		err := m.Validate()
		if err != nil {
			return nil, err
		}
	}
	err = options.Validate()
	if err != nil {
		return nil, err
	}

	if options.RagQuery.Input != "" {
		messages, err = a.EnrichWithRag(ctx, messages, options.RagQuery)
		if err != nil {
			return nil, err
		}
	}

	fullMessages, err := a.Enrich(ctx, context, messages)
	if err != nil {
		return nil, err
	}
	answer, err := a.Message(ctx, fullMessages, options)
	if err != nil {
		return nil, err
	}
	answer.Context = context.ID
	err = a.UpdateContext(ctx, context.ID, messages, answer.Results)
	if err != nil {
		return nil, err
	}
	return answer, nil
}
