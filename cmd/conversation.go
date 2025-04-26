package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/spf13/cobra"
)

func buildConversationCmd() *cobra.Command {
	var sourcesContextID []string
	var sourcesContextName []string
	var model string
	var messages []string
	var fileMessages []string
	var aiProvider string
	var system string
	var temperature float64
	var maxTokens uint64
	var contextID string
	var newContextName string
	var contextName string
	var newContextDescription string
	var interactive bool
	var stream bool
	var ragInput string
	var ragModel string
	var ragProvider string
	var ragLimit uint32
	cmd := &cobra.Command{
		Use: "conversation",
		Short: `Send a message to an AI provider.
If a context ID is provided, it will be used as input for the conversation. Else, a context name should be provided.`,
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			options := client.QueryOptions{
				Model:       model,
				System:      system,
				Temperature: temperature,
				MaxTokens:   maxTokens,
				Provider:    aiProvider,
				RagQuery: client.RagSearchQuery{
					Input:    ragInput,
					Provider: ragProvider,
					Model:    ragModel,
					Limit:    int32(ragLimit),
				},
			}
			if contextID != "" && contextName != "" {
				exitIfError(errors.New("You shoulh pass either a context ID or a context name"))
			}
			if contextName != "" {
				context, err := c.GetContextByName(ctx, contextName)
				exitIfError(err)
				contextID = context.ID
			}
			if newContextName == "" && contextID == "" {
				newContextName = fmt.Sprintf("context-auto-%d", time.Now().Unix())
			}
			for _, ctxName := range sourcesContextName {
				context, err := c.GetContextByName(ctx, ctxName)
				exitIfError(err)
				sourcesContextID = append(sourcesContextID, context.ID)
			}
			contextOptions := client.ContextOptions{
				Name:        newContextName,
				Description: newContextDescription,
				Sources: client.ContextSources{
					Contexts: sourcesContextID,
				},
			}
			msg := toMessages(messages)
			for _, input := range fileMessages {
				role, path, found := strings.Cut(input, ":")
				if !found {
					exitIfError(errors.New("files paths should start with the role to use"))
				}
				content, err := os.ReadFile(path)
				if err != nil {
					exitIfError(fmt.Errorf("fail to read file %s: %w", path, err))
				}
				msg = append(msg, client.NewMessage{
					Role:    role,
					Content: string(content),
				})
			}
			input := &client.CreateConversationInput{
				QueryOptions:      options,
				NewContextOptions: contextOptions,
				Messages:          msg,
			}
			if interactive {
				input.Stream = stream
				fmt.Printf("Hello, I'm your AI assistant. Ask me anything:\n\n")
				reader := bufio.NewReader(os.Stdin)
				var updatedContextID = contextID
				for {
					input.ContextID = updatedContextID
					prompt, err := reader.ReadString('\n')
					prompt = strings.TrimSpace(prompt)
					fmt.Println("")
					exitIfError(err)
					if prompt == "exit" || prompt == "exit\n" {
						return
					}
					input.Messages = []client.NewMessage{
						{
							Role:    "user",
							Content: prompt,
						},
					}
					if stream {
						eventChan, err := c.StreamConversation(ctx, *input)
						exitIfError(err)
						for event := range eventChan {
							fmt.Print(event.Delta)
							if event.Error != "" {
								fmt.Printf("\nerror: %s\n", event.Error)
							}
							if event.InputTokens != 0 {
								fmt.Printf("\n\nInput tokens: %d, output tokens: %d\n", event.InputTokens, event.OutputTokens)
							}
							if event.Context != "" {
								updatedContextID = event.Context
							}
						}

					} else {
						answer, err := c.CreateConversation(ctx, *input)
						exitIfError(err)
						fmt.Printf("\nAnswer (input tokens %d, output tokens %d):\n\n", answer.InputTokens, answer.OutputTokens)
						for _, result := range answer.Results {
							fmt.Printf("\n%s\n", result.Text)
						}
						updatedContextID = answer.Context
					}
					fmt.Printf("\nAnything else (write 'exit' to exit the program)?\n\n")
				}

			} else {
				input.ContextID = contextID
				answer, err := c.CreateConversation(ctx, *input)
				exitIfError(err)
				printJson(answer)
			}

		},
	}

	cmd.PersistentFlags().StringVar(&model, "model", "", "Model to use")
	err := cmd.MarkPersistentFlagRequired("model")
	exitIfError(err)

	cmd.PersistentFlags().StringArrayVar(&messages, "message", []string{}, "The messages to send to the AI provider. You can use the {maizai_rag_data} placeholder: it will be replaced by RAG data if a rag input is provided")
	cmd.PersistentFlags().StringArrayVar(&fileMessages, "message-from-file", []string{}, "A list of files paths, the content will be added to the context. They should be prefixed by the role name (example: user:/my/file)")

	cmd.PersistentFlags().StringVar(&aiProvider, "provider", "", "AI provider to use")
	err = cmd.MarkPersistentFlagRequired("provider")
	exitIfError(err)

	cmd.PersistentFlags().StringVar(&system, "system", "", "System promt for the AI provider")
	cmd.PersistentFlags().StringVar(&contextID, "context-id", "", "The ID of the context to reuse for this conversation")
	cmd.PersistentFlags().StringVar(&contextName, "context-name", "", "The name of the context to reuse for this conversation")
	cmd.PersistentFlags().StringVar(&newContextName, "new-context-name", "", "The name of the new context that will be created for this conversation if a context ID is not provided")
	cmd.PersistentFlags().StringVar(&newContextDescription, "new-context-description", "", "The description of the new context that will be created for this conversation if a context ID is not provided")
	cmd.PersistentFlags().Float64Var(&temperature, "temperature", 0, "Temperature")
	cmd.PersistentFlags().Uint64Var(&maxTokens, "max-tokens", 8192, "Maximum tokens on the answer")
	cmd.PersistentFlags().StringArrayVar(&sourcesContextID, "source-context-id", []string{}, "ID of a context to use as source")
	cmd.PersistentFlags().StringArrayVar(&sourcesContextName, "source-context-name", []string{}, "name of a context to use as source")
	cmd.PersistentFlags().BoolVar(&interactive, "interactive", false, "Starts an interactive conversation")
	cmd.PersistentFlags().BoolVar(&stream, "stream", false, "Streams the conversation")
	cmd.PersistentFlags().StringVar(&ragInput, "rag-input", "", "Input to use to fetch data from MaiZAI RAG")
	cmd.PersistentFlags().StringVar(&ragModel, "rag-model", "mistral-embed", "Model to use for the rag")
	cmd.PersistentFlags().StringVar(&ragProvider, "rag-provider", "mistral", "The AI provider to use for the rag")
	cmd.PersistentFlags().Uint32Var(&ragLimit, "rag-limit", 1, "The number of chunks to return from the RAG to enrich the context")
	exitIfError(err)
	return cmd
}
