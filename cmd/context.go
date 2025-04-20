package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/spf13/cobra"
)

func contextListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contexts",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.ListContexts(ctx)
			exitIfError(err)
			printJson(contexts)
		},
	}
	return cmd
}

func contextGetCmd() *cobra.Command {
	var id string
	var name string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a context by ID or name",
		Run: func(cmd *cobra.Command, args []string) {
			if id == "" && name == "" {
				exitIfError(errors.New("the command expects either a context id or name as input"))
			}
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			if id != "" {
				context, err := client.GetContext(ctx, id)
				exitIfError(err)
				printJson(*context)
			} else {
				contexts, err := client.ListContexts(ctx)
				exitIfError(err)
				for _, context := range contexts.Contexts {
					if context.Name == name {
						printJson(context)
						return
					}
				}
				exitIfError(fmt.Errorf("context %s not found", name))
			}

		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the context to retrieve")
	cmd.PersistentFlags().StringVar(&name, "name", "", "The name of the context to retrieve")
	return cmd
}

func contextDeleteCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a context by ID",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			response, err := client.DeleteContext(ctx, id)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the context to delete")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func contextSourceContextDeleteCmd() *cobra.Command {
	var id string
	var sourceContextID string
	cmd := &cobra.Command{
		Use:   "delete-context",
		Short: "Delete a context used as a source in a context",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.DeleteContextSourceContextInput{
				ID:              id,
				SourceContextID: sourceContextID,
			}
			response, err := c.DeleteContextSourceContext(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the context")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&sourceContextID, "source-context-id", "", "The ID of the source context")
	err = cmd.MarkPersistentFlagRequired("source-context-id")
	exitIfError(err)
	return cmd
}

func contextSourceContextAddCmd() *cobra.Command {
	var id string
	var sourceContextID string
	cmd := &cobra.Command{
		Use:   "add-context",
		Short: "Add a context as a source in another context",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.CreateContextSourceContextInput{
				ID:              id,
				SourceContextID: sourceContextID,
			}
			response, err := c.CreateContextSourceContext(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the context")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&sourceContextID, "source-context-id", "", "The ID of the source context")
	err = cmd.MarkPersistentFlagRequired("source-context-id")
	exitIfError(err)
	return cmd
}

func toMessage(input string) client.NewMessage {
	role, content, found := strings.Cut(input, ":")
	if !found {
		exitIfError(errors.New("Contexts messages should start with the role to use for this message"))
	}
	return client.NewMessage{
		Role:    role,
		Content: content,
	}
}

func toMessages(inputs []string) []client.NewMessage {
	result := []client.NewMessage{}
	for _, message := range inputs {
		result = append(result, toMessage(message))
	}
	return result
}

func addMessagesToContextCmd() *cobra.Command {
	var id string
	var messages []string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add messages to a given context",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.AddMessagesToContextInput{
				ID:       id,
				Messages: toMessages(messages),
			}
			response, err := c.AddMessagesToContext(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The context ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	cmd.PersistentFlags().StringArrayVar(&messages, "message", []string{}, "Messages to add to this context")
	return cmd
}

func deleteContextMessageCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a messages For a given context",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.DeleteContextMessageInput{
				ID: id,
			}
			response, err := c.DeleteContextMessage(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The context message ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func messageUpdateCmd() *cobra.Command {
	var id string
	var message string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a context message",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			msg := toMessage(message)
			input := client.UpdateContextMessageInput{
				ID:      id,
				Role:    msg.Role,
				Content: msg.Content,
			}
			response, err := c.UpdateContextMessage(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "The ID of the context to delete")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&message, "message", "", "The new message role and content")
	err = cmd.MarkPersistentFlagRequired("message")
	exitIfError(err)
	return cmd
}

func contextCreateCmd() *cobra.Command {
	var name string
	var description string
	var sourcesContext []string
	var messages []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new context",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.CreateContextInput{
				Name:        name,
				Description: description,
				Sources: shared.ContextSources{
					Contexts: sourcesContext,
				},
				Messages: toMessages(messages),
			}
			response, err := c.CreateContext(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", "", "The name of the new context")
	err := cmd.MarkPersistentFlagRequired("name")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&description, "description", "", "The description of the new context")
	cmd.PersistentFlags().StringArrayVar(&sourcesContext, "source-context", []string{}, "IDs of contexts to use as source for this context")
	cmd.PersistentFlags().StringArrayVar(&messages, "message", []string{}, "Messages to add to this context")
	return cmd
}

// func contextCreateFromFilesCmd() *cobra.Command {
// 	var files []string
// 	var directories []string
// 	var name string
// 	cmd := &cobra.Command{
// 		Use:   "files",
// 		Short: "Creates a new context from local files",
// 		Run: func(cmd *cobra.Command, args []string) {

// 			config, err := config.Load()
// 			exitIfError(err)
// 			store, err := buildContextStore(config.Store)
// 			exitIfError(err)
// 			manager := maizaictx.New(store)
// 			localFilesContext := caggregates.LocalFileContext{
// 				Files:       files,
// 				Directories: []caggregates.Directory{},
// 			}
// 			for _, directory := range directories {
// 				d := caggregates.Directory{
// 					Path:      directory,
// 					Recursive: false,
// 				}
// 				localFilesContext.Directories = append(localFilesContext.Directories, d)
// 			}
// 			options := shared.ContextOptions{
// 				Name: name,
// 			}
// 			err = manager.FromFiles(context.Background(), localFilesContext, options)
// 			exitIfError(err)
// 		},
// 	}
// 	cmd.PersistentFlags().StringArrayVar(&files, "file", []string{}, "File to include in the context")
// 	cmd.PersistentFlags().StringArrayVar(&directories, "directory", []string{}, "Directory to include in the context")

// 	cmd.PersistentFlags().StringVar(&name, "name", "", "The name of the new context")
// 	err := cmd.MarkPersistentFlagRequired("name")
// 	exitIfError(err)

// 	return cmd
// }
