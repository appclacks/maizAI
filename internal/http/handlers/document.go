package handlers

import (
	"net/http"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/labstack/echo/v4"
)

func toClientDocument(document aggregates.Document) client.Document {
	return client.Document{
		ID:          document.ID,
		Name:        document.Name,
		Description: document.Description,
		CreatedAt:   document.CreatedAt,
	}
}

func toClientDocumentChunk(chunk aggregates.DocumentChunk) client.DocumentChunk {
	return client.DocumentChunk{
		ID:         chunk.ID,
		DocumentID: chunk.DocumentID,
		Fragment:   chunk.Fragment,
		Embedding:  chunk.Embedding,
		CreatedAt:  chunk.CreatedAt,
	}
}

func (b *Builder) ListDocuments(ec echo.Context) error {
	documents, err := b.ragManager.ListDocuments(ec.Request().Context())
	if err != nil {
		return err
	}
	docs := []client.Document{}
	for _, d := range documents {
		docs = append(docs, toClientDocument(d))
	}
	return ec.JSON(http.StatusOK, client.ListDocumentsOutput{
		Documents: docs,
	})
}

func (b *Builder) CreateDocument(ec echo.Context) error {
	var payload client.CreateDocumentInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	document, err := aggregates.NewDocument(payload.Name, payload.Description)
	if err != nil {
		return err
	}
	err = b.ragManager.CreateDocument(ec.Request().Context(), *document)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("document created"))
}

func (b *Builder) EmbedDocument(ec echo.Context) error {
	var payload client.EmbedDocumentInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	query := aggregates.EmbeddingQuery{
		Model:    payload.Model,
		Input:    payload.Input,
		Provider: payload.Provider,
	}
	err := b.ragManager.Embed(ec.Request().Context(), payload.DocumentID, query)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("document chunk created"))
}

func (b *Builder) GetDocument(ec echo.Context) error {
	var payload client.GetDocumentInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	doc, err := b.ragManager.GetDocument(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, toClientDocument(*doc))
}

func (b *Builder) DeleteDocument(ec echo.Context) error {
	var payload client.DeleteDocumentInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ragManager.DeleteDocument(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("document deleted"))
}

func (b *Builder) DeleteDocumentChunk(ec echo.Context) error {
	var payload client.DeleteDocumentChunkInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	err := b.ragManager.DeleteDocumentChunk(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, newResponse("document chunk deleted"))
}

func (b *Builder) MatchChunk(ec echo.Context) error {
	var payload client.RagSearchQuery
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	chunks, err := b.ragManager.Match(ec.Request().Context(), aggregates.SearchQuery{
		Input:    payload.Input,
		Model:    payload.Model,
		Provider: payload.Provider,
		Limit:    payload.Limit,
	})
	if err != nil {
		return err
	}
	response := client.ListDocumentChunksOutput{
		Chunks: []client.DocumentChunk{},
	}
	for _, c := range chunks {
		response.Chunks = append(response.Chunks, toClientDocumentChunk(c))
	}
	return ec.JSON(http.StatusOK, response)
}

func (b *Builder) ListDocumentChunksForDocument(ec echo.Context) error {
	var payload client.ListDocumentChunksForDocumentInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	chunks, err := b.ragManager.ListDocumentChunksForDocument(ec.Request().Context(), payload.DocumentID)
	if err != nil {
		return err
	}
	response := client.ListDocumentChunksOutput{
		Chunks: []client.DocumentChunk{},
	}
	for _, c := range chunks {
		response.Chunks = append(response.Chunks, toClientDocumentChunk(c))
	}
	return ec.JSON(http.StatusOK, response)
}
