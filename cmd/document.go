package cmd

import (
	"context"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/spf13/cobra"
)

func documentListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.ListDocuments(ctx)
			exitIfError(err)
			printJson(contexts)
		},
	}
	return cmd
}

func documentCreateCmd() *cobra.Command {
	var name string
	var description string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new document",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.CreateDocumentInput{
				Name:        name,
				Description: description,
			}
			response, err := c.CreateDocument(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", "", "The name of the new document")
	err := cmd.MarkPersistentFlagRequired("name")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&description, "description", "", "The description of the new document")
	return cmd
}

func documentEmbedCmd() *cobra.Command {
	var input string
	var model string
	var aiProvider string
	var docID string
	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Embed a chunk for a specific document, that will be stored in MaiZAI RAG",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.EmbedDocumentInput{
				DocumentID: docID,
				Model:      model,
				Input:      input,
				Provider:   aiProvider,
			}
			response, err := c.EmbedDocument(ctx, input)
			exitIfError(err)
			printJson(*response)
		},
	}
	cmd.PersistentFlags().StringVar(&input, "input", "", "Content to embed")
	err := cmd.MarkPersistentFlagRequired("input")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&docID, "document-id", "", "The ID of the document linked to this content")
	err = cmd.MarkPersistentFlagRequired("document-id")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&model, "model", "mistral-embed", "The model to use")
	cmd.PersistentFlags().StringVar(&aiProvider, "provider", "mistral", "The AI provider to use")
	return cmd
}

func documentGetCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a document by ID",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.GetDocument(ctx, id)
			exitIfError(err)
			printJson(contexts)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "Document ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func documentDeleteCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a document by ID",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.DeleteDocument(ctx, id)
			exitIfError(err)
			printJson(contexts)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "Document ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func embeddingMatchCmd() *cobra.Command {
	var limit int32
	var input string
	var model string
	var aiProvider string
	cmd := &cobra.Command{
		Use:   "match",
		Short: "Get closest chunks in MAizAI RAG for the input",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			input := client.RagSearchQuery{
				Input:    input,
				Model:    model,
				Provider: aiProvider,
				Limit:    limit,
			}
			contexts, err := c.MatchChunk(ctx, input)
			exitIfError(err)
			printJson(contexts)
		},
	}
	cmd.PersistentFlags().StringVar(&input, "input", "", "Content to use as query for the rag")
	err := cmd.MarkPersistentFlagRequired("input")
	exitIfError(err)
	cmd.PersistentFlags().StringVar(&model, "model", "mistral-embed", "The model to use")
	cmd.PersistentFlags().StringVar(&aiProvider, "provider", "mistral", "The AI provider to use")
	cmd.PersistentFlags().Int32Var(&limit, "limit", 1, "Number of chunks to return")
	return cmd
}

func documentChunkListCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "list-chunks",
		Short: "List chunks for a document",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.ListDocumentsChunkForDocument(ctx, id)
			exitIfError(err)
			printJson(contexts)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "Document ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}

func documentChunkDeleteCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Get a document chunk by ID",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := client.New()
			exitIfError(err)
			ctx := context.Background()
			contexts, err := client.DeleteDocumentChunk(ctx, id)
			exitIfError(err)
			printJson(contexts)
		},
	}
	cmd.PersistentFlags().StringVar(&id, "id", "", "Document chunk ID")
	err := cmd.MarkPersistentFlagRequired("id")
	exitIfError(err)
	return cmd
}
