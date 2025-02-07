package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/spf13/cobra"
)

func buildConversationCmd() *cobra.Command {
	var sourcesContext []string
	var model string
	var prompt string
	var aiProvider string
	var system string
	var temperature float64
	var maxTokens uint64
	var contextID string
	var contextName string
	var contextDescription string
	var interactive bool
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
			contextOptions := client.ContextOptions{
				Name:        contextName,
				Description: contextDescription,
				Sources: client.ContextSources{
					Contexts: sourcesContext,
				},
			}

			input := &client.CreateConversationInput{
				QueryOptions:      options,
				NewContextOptions: contextOptions,
				Prompt:            prompt,
			}
			if interactive {
				fmt.Printf("Hello, I'm your AI assistant. Ask me anything:\n\n")
				reader := bufio.NewReader(os.Stdin)
				var updatedContextID = contextID
				for {
					input.ContextID = updatedContextID
					prompt, err := reader.ReadString('\n')
					exitIfError(err)
					if prompt == "exit" || prompt == "exit\n" {
						return
					}
					input.Prompt = prompt
					answer, err := c.CreateConversation(ctx, *input)
					exitIfError(err)
					fmt.Printf("\nAnswer (input tokens %d, output tokens %d):\n\n", answer.InputTokens, answer.OutputTokens)
					for _, result := range answer.Results {
						fmt.Printf("%s\n", result.Text)
					}
					updatedContextID = answer.Context
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

	cmd.PersistentFlags().StringVar(&prompt, "prompt", "", "The prompt to send to the AI provider. You can use the {maizai_rag_data} placeholder: it will be replaced by RAG data if a rag input is provided")

	cmd.PersistentFlags().StringVar(&aiProvider, "provider", "", "AI provider to use")
	err = cmd.MarkPersistentFlagRequired("provider")
	exitIfError(err)

	cmd.PersistentFlags().StringVar(&system, "system", "", "System promt for the AI provider")
	cmd.PersistentFlags().StringVar(&contextID, "context-id", "", "The ID of the new context to reuse for this conversation")
	cmd.PersistentFlags().StringVar(&contextName, "context-name", "", "The name of the new context that will be created for this conversation if a context ID is not provided")
	cmd.PersistentFlags().Float64Var(&temperature, "temperature", 0, "Temperature")
	cmd.PersistentFlags().Uint64Var(&maxTokens, "max-tokens", 8192, "Maximum tokens on the answer")
	cmd.PersistentFlags().StringSliceVar(&sourcesContext, "source-context", []string{}, "Contexts to load")
	cmd.PersistentFlags().BoolVar(&interactive, "interactive", false, "Starts an interactive conversation")

	cmd.PersistentFlags().StringVar(&ragInput, "rag-input", "", "Input to use to fetch data from MaiZAI RAG")
	cmd.PersistentFlags().StringVar(&ragModel, "rag-model", "mistral-embed", "Model to use for the rag")
	cmd.PersistentFlags().StringVar(&ragProvider, "rag-provider", "mistral", "The AI provider to use for the rag")
	cmd.PersistentFlags().Uint32Var(&ragLimit, "rag-limit", 1, "The number of chunks to return from the RAG to enrich the context")
	exitIfError(err)
	return cmd
}
