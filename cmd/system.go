package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/spf13/cobra"
)

func systemPromptListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List system prompts",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			prompts, err := client.ListSystemPrompts(ctx)
			exitIfError(err)
			printJson(prompts)
		},
	}
	return cmd
}

func systemPromptGetCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a system prompt by ID",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" {
				exitIfError(errors.New("the command expects a system prompt id as input"))
			}
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			prompt, err := client.GetSystemPrompt(ctx, id)
			exitIfError(err)
			printJson(*prompt)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the system prompt to retrieve")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func systemPromptCreateCmd() *cobra.Command {
	var name string
	var description string
	var content string
	var contentFile string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new system prompt",
		Run: func(cmd *cobra.Command, args []string) {
			if content == "" && contentFile == "" {
				exitIfError(errors.New("either --content or --content-from-file must be provided"))
			}
			if content != "" && contentFile != "" {
				exitIfError(errors.New("cannot specify both --content and --content-from-file"))
			}

			finalContent := content
			if contentFile != "" {
				fileContent, err := os.ReadFile(contentFile)
				if err != nil {
					exitIfError(err)
				}
				finalContent = string(fileContent)
			}

			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.CreateSystemPromptInput{
				Name:        name,
				Description: description,
				Content:     finalContent,
			}
			response, err := c.CreateSystemPrompt(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", "", "The name of the new system prompt")
	err := cmd.MarkPersistentFlagRequired("name")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&description, "description", "", "The description of the new system prompt")
	cmd.PersistentFlags().StringVar(&content, "content", "", "The content of the system prompt")
	cmd.PersistentFlags().StringVar(&contentFile, "content-from-file", "", "Path to file containing the system prompt content")
	return cmd
}

func systemPromptUpdateCmd() *cobra.Command {
	var id string
	var content string
	var contentFile string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a system prompt's content",
		Run: func(cmd *cobra.Command, args []string) {
			if content == "" && contentFile == "" {
				exitIfError(errors.New("either --content or --content-from-file must be provided"))
			}
			if content != "" && contentFile != "" {
				exitIfError(errors.New("cannot specify both --content and --content-from-file"))
			}

			finalContent := content
			if contentFile != "" {
				fileContent, err := os.ReadFile(contentFile)
				if err != nil {
					exitIfError(err)
				}
				finalContent = string(fileContent)
			}

			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.UpdateSystemPromptInput{
				ID:      id,
				Content: finalContent,
			}
			response, err := c.UpdateSystemPrompt(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the system prompt to update")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&content, "content", "", "The new content for the system prompt")
	cmd.PersistentFlags().StringVar(&contentFile, "content-from-file", "", "Path to file containing the new system prompt content")
	return cmd
}

func systemPromptDeleteCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a system prompt by ID",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			response, err := client.DeleteSystemPrompt(ctx, id)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the system prompt to delete")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}