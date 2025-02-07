package handlers

import (
	"net/http"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/context"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/labstack/echo/v4"
)

func toClientMetadata(context shared.ContextMetadata) client.ContextMetadata {
	result := client.ContextMetadata{
		ID:          context.ID,
		Name:        context.Name,
		Description: context.Description,
		Sources: client.ContextSources{
			Contexts: context.Sources.Contexts,
		},
		CreatedAt: context.CreatedAt,
	}
	return result
}

func toClientContext(context shared.Context) client.Context {
	result := client.Context{
		ID:          context.ID,
		Name:        context.Name,
		Description: context.Description,
		Sources: client.ContextSources{
			Contexts: context.Sources.Contexts,
		},
		Messages:  []client.Message{},
		CreatedAt: context.CreatedAt,
	}
	for _, message := range context.Messages {
		result.Messages = append(result.Messages, client.Message{
			ID:        message.ID,
			Role:      message.Role,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		})
	}
	return result
}

func (b *Builder) ListContexts(ec echo.Context) error {
	contexts, err := b.ctxManager.ListContexts(ec.Request().Context())
	if err != nil {
		return err
	}
	output := client.ListContextOutput{
		Contexts: []client.ContextMetadata{},
	}
	for _, context := range contexts {
		output.Contexts = append(output.Contexts, toClientMetadata(context))
	}
	return ec.JSON(http.StatusOK, output)
}

func (b *Builder) GetContext(ec echo.Context) error {
	var payload client.GetContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	context, err := b.ctxManager.GetContext(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, toClientContext(*context))
}

func (b *Builder) CreateContext(ec echo.Context) error {
	var payload client.CreateContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	options := shared.ContextOptions{
		Name:        payload.Name,
		Description: payload.Description,
		Sources:     payload.Sources,
	}
	context, err := context.NewContext(options)
	if err != nil {
		return err
	}
	for _, message := range payload.Messages {
		newMessage, err := shared.NewMessage(message.Role, message.Content)
		if err != nil {
			return err
		}
		context.Messages = append(context.Messages, *newMessage)

	}
	err = b.ctxManager.CreateContext(ec.Request().Context(), *context)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("context created"))
}

func (b *Builder) DeleteContext(ec echo.Context) error {
	var payload client.DeleteContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ctxManager.DeleteContext(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("context deleted"))
}

func (b *Builder) AddMessagesToContext(ec echo.Context) error {
	var payload client.AddMessagesToContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	messages := []shared.Message{}
	for _, message := range payload.Messages {
		msg, err := shared.NewMessage(message.Role, message.Content)
		if err != nil {
			return err
		}
		messages = append(messages, *msg)
	}
	err := b.ctxManager.AddMessagesToContext(ec.Request().Context(), payload.ID, messages)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("messages added to context"))
}

func (b *Builder) DeleteContextMessage(ec echo.Context) error {
	var payload client.DeleteContextMessageInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ctxManager.DeleteContextMessage(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("message deleted from context"))
}

func (b *Builder) UpdateContextMessage(ec echo.Context) error {
	var payload client.UpdateContextMessageInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ctxManager.UpdateContextMessage(ec.Request().Context(), payload.ID, payload.Role, payload.Content)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("message updated"))
}

func (b *Builder) DeleteContextSourceContext(ec echo.Context) error {
	var payload client.DeleteContextSourceContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ctxManager.DeleteContextSourceContext(ec.Request().Context(), payload.ID, payload.SourceContextID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("Context source deleted"))
}

func (b *Builder) CreateContextSourceContext(ec echo.Context) error {
	var payload client.CreateContextSourceContextInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ctxManager.CreateContextSourceContext(ec.Request().Context(), payload.ID, payload.SourceContextID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("Context source added"))
}
