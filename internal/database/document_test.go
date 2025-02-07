package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDocumentCRUD(t *testing.T) {
	ctx := context.Background()
	doc := aggregates.Document{
		ID:          uuid.NewString(),
		Name:        "doc1",
		CreatedAt:   time.Now().UTC(),
		Description: "desc1",
	}
	err := TestComponent.CreateDocument(ctx, doc)
	assert.NoError(t, err)

	retrieved, err := TestComponent.GetDocument(ctx, doc.ID)
	assert.NoError(t, err)
	assert.Equal(t, doc.ID, retrieved.ID)
	assert.Equal(t, doc.Name, retrieved.Name)
	assert.Equal(t, doc.Description, retrieved.Description)

	doc2 := aggregates.Document{
		ID:          uuid.NewString(),
		Name:        "doc2",
		CreatedAt:   time.Now().UTC(),
		Description: "desc2",
	}
	err = TestComponent.CreateDocument(ctx, doc2)
	assert.NoError(t, err)

	list, err := TestComponent.ListDocuments(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)

	err = TestComponent.DeleteDocument(ctx, doc.ID)
	assert.NoError(t, err)
	_, err = TestComponent.GetDocument(ctx, doc.ID)
	assert.ErrorContains(t, err, "doesn't exist")

	list, err = TestComponent.ListDocuments(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	embedding := []float32{}
	for i := 0; i < 1024; i++ {
		embedding = append(embedding, float32(i))
	}
	chunk := aggregates.DocumentChunk{
		ID:         uuid.NewString(),
		DocumentID: doc2.ID,
		Fragment:   "hello world",
		Embedding:  embedding,
	}
	err = TestComponent.CreateDocumentChunk(ctx, chunk)
	assert.NoError(t, err)

	chunks, err := TestComponent.ListDocumentChunksForDocument(ctx, doc2.ID)
	assert.NoError(t, err)
	assert.Len(t, chunks, 1)
	assert.Equal(t, chunk.ID, chunks[0].ID)
	assert.Equal(t, chunk.DocumentID, chunks[0].DocumentID)
	assert.Equal(t, chunk.Fragment, chunks[0].Fragment)

	err = TestComponent.DeleteDocumentChunk(ctx, chunk.ID)
	assert.NoError(t, err)

	chunks, err = TestComponent.ListDocumentChunksForDocument(ctx, doc2.ID)
	assert.NoError(t, err)
	assert.Len(t, chunks, 0)

	_, err = TestComponent.ListDocumentChunksForDocument(ctx, doc.ID)
	assert.ErrorContains(t, err, "doesn't exist")
}
