package handlers

import (
	"net/http"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/labstack/echo/v4"
)

func toClientSystemPrompt(prompt shared.SystemPrompt) client.SystemPrompt {
	return client.SystemPrompt{
		ID:          prompt.ID,
		Name:        prompt.Name,
		Description: prompt.Description,
		Content:     prompt.Content,
		CreatedAt:   prompt.CreatedAt,
	}
}

func (b *Builder) ListSystemPrompts(ec echo.Context) error {
	prompts, err := b.systemPromptManager.ListSystemPrompts(ec.Request().Context())
	if err != nil {
		return err
	}
	output := client.ListSystemPromptsOutput{
		SystemPrompts: []client.SystemPrompt{},
	}
	for _, prompt := range prompts {
		output.SystemPrompts = append(output.SystemPrompts, toClientSystemPrompt(prompt))
	}
	return ec.JSON(http.StatusOK, output)
}

func (b *Builder) GetSystemPrompt(ec echo.Context) error {
	var payload client.GetSystemPromptInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	prompt, err := b.systemPromptManager.GetSystemPrompt(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, toClientSystemPrompt(*prompt))
}

func (b *Builder) CreateSystemPrompt(ec echo.Context) error {
	var payload client.CreateSystemPromptInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	prompt, err := shared.NewSystemPrompt(payload.Name, payload.Description, payload.Content)
	if err != nil {
		return err
	}
	err = b.systemPromptManager.CreateSystemPrompt(ec.Request().Context(), *prompt)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("system prompt created"))
}

func (b *Builder) UpdateSystemPrompt(ec echo.Context) error {
	var payload client.UpdateSystemPromptInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.systemPromptManager.UpdateSystemPrompt(ec.Request().Context(), payload.ID, payload.Content)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("system prompt updated"))
}

func (b *Builder) DeleteSystemPrompt(ec echo.Context) error {
	var payload client.DeleteSystemPromptInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.systemPromptManager.DeleteSystemPrompt(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("system prompt deleted"))
}
