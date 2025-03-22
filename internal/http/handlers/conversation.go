package handlers

import (
	"encoding/json"
	"fmt"
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
	if payload.Stream {
		eventChan, err := b.assistant.StreamPipeline(ctx, queryOpts, contextOpts, payload.ContextID, messages)
		if err != nil {
			return err
		}

		w := ec.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		for event := range eventChan {
			e := client.ConversationStreamEvent{
				Delta: event.Delta,
			}
			if event.Error != nil {
				e.Error = event.Error.Error()
			}
			if event.Answer != nil {
				e.InputTokens = event.Answer.InputTokens
				e.OutputTokens = event.Answer.OutputTokens
				e.Context = event.Answer.Context
			}
			j, err := json.Marshal(e)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "data: %s\n", j)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(w, "\n")
			if err != nil {
				return err
			}
			w.Flush()

		}
		return nil
	} else {
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
}
