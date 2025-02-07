package handlers

import (
	"net/http"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/assistant/aggregates"
	ragdata "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/labstack/echo/v4"
)

func (b *Builder) Conversation(ec echo.Context) error {
	var payload client.CreateConversationInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}

	messages, err := shared.NewUserMessages(payload.Prompt)
	if err != nil {
		return err
	}
	ctx := ec.Request().Context()
	queryOpts := aggregates.QueryOptions{
		Model:       payload.QueryOptions.Model,
		System:      payload.QueryOptions.System,
		Temperature: payload.QueryOptions.Temperature,
		MaxTokens:   payload.QueryOptions.MaxTokens,
		Provider:    payload.QueryOptions.Provider,
		RagQuery: ragdata.SearchQuery{
			Input:    payload.QueryOptions.RagQuery.Input,
			Model:    payload.QueryOptions.RagQuery.Model,
			Provider: payload.QueryOptions.RagQuery.Provider,
			Limit:    payload.QueryOptions.RagQuery.Limit,
		},
	}
	contextOpts := shared.ContextOptions{
		Name:        payload.NewContextOptions.Name,
		Description: payload.NewContextOptions.Description,
		Sources: shared.ContextSources{
			Contexts: payload.NewContextOptions.Sources.Contexts,
		},
	}
	answer, err := b.assistant.Pipeline(ctx, queryOpts, contextOpts, payload.ContextID, messages)
	if err != nil {
		return err
	}
	response := client.ConversationAnswer{
		Results:      []client.Result{},
		InputTokens:  answer.InputTokens,
		OutputTokens: answer.OutputTokens,
		Context:      answer.Context,
	}
	for _, result := range answer.Results {
		response.Results = append(response.Results, client.Result{
			Text: result.Text,
		})
	}
	return ec.JSON(http.StatusOK, response)
}
