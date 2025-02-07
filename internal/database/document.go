package database

import (
	"context"

	"github.com/appclacks/maizai/internal/database/queries"
	"github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/jackc/pgx/v5"
	er "github.com/mcorbin/corbierror"
	"github.com/pgvector/pgvector-go"
)

func (c *Database) CreateDocument(ctx context.Context, document aggregates.Document) error {
	err := c.queries.CreateDocument(ctx, queries.CreateDocumentParams{
		ID:          pgxID(document.ID),
		Name:        document.Name,
		Description: pgxText(document.Description),
		CreatedAt:   pgxTime(document.CreatedAt),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) GetDocument(ctx context.Context, id string) (*aggregates.Document, error) {
	document, err := c.queries.GetDocument(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
		return nil, er.Newf("document %s doesn't exist", er.NotFound, true, id)
	}

	return &aggregates.Document{
		ID:          document.ID.String(),
		Name:        document.Name,
		Description: document.Description.String,
		CreatedAt:   document.CreatedAt.Time,
	}, nil
}

func (c *Database) ListDocuments(ctx context.Context) ([]aggregates.Document, error) {
	documents, err := c.queries.ListDocuments(ctx)
	if err != nil {
		return nil, err
	}
	result := []aggregates.Document{}

	for _, document := range documents {
		doc := aggregates.Document{
			ID:          document.ID.String(),
			Name:        document.Name,
			Description: document.Description.String,
			CreatedAt:   document.CreatedAt.Time,
		}
		result = append(result, doc)
	}
	return result, nil
}

func (c *Database) DeleteDocument(ctx context.Context, id string) error {
	tx, qtx, rollbackFn, err := c.beginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackFn()
	err = qtx.DeleteDocumentChunkForDocument(ctx, pgxID(id))
	if err != nil {
		return err
	}
	err = qtx.DeleteDocument(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return err
		}
		return er.Newf("document %s doesn't exist", er.NotFound, true, id)
	}
	return tx.Commit(ctx)
}

func (c *Database) documentExists(queries *queries.Queries, ctx context.Context, id string) (bool, error) {
	_, err := queries.GetDocument(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (c *Database) CreateDocumentChunk(ctx context.Context, documentChunk aggregates.DocumentChunk) error {
	err := c.queries.CreateDocumentChunk(ctx, queries.CreateDocumentChunkParams{
		ID:         pgxID(documentChunk.ID),
		DocumentID: pgxID(documentChunk.DocumentID),
		Fragment:   pgxText(documentChunk.Fragment),
		Embedding:  pgvector.NewVector(documentChunk.Embedding),
		CreatedAt:  pgxTime(documentChunk.CreatedAt),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) FindClosestChunks(ctx context.Context, limit int32, chunk []float32) ([]aggregates.DocumentChunk, error) {
	chunks, err := c.queries.FindClosestChunks(ctx, queries.FindClosestChunksParams{
		Limit:     limit,
		Embedding: pgvector.NewVector(chunk),
	})
	if err != nil {
		return nil, err
	}
	result := []aggregates.DocumentChunk{}
	for _, chunk := range chunks {
		result = append(result, aggregates.DocumentChunk{
			ID:         chunk.ID.String(),
			DocumentID: chunk.DocumentID.String(),
			Fragment:   chunk.Fragment.String,
			Embedding:  chunk.Embedding.Slice(),
			CreatedAt:  chunk.CreatedAt.Time,
		})
	}
	return result, nil
}

func (c *Database) DeleteDocumentChunk(ctx context.Context, id string) error {
	err := c.queries.DeleteDocumentChunk(ctx, pgxID(id))
	if err != nil {
		if err != pgx.ErrNoRows {
			return err
		}
		return er.Newf("document chunk %s doesn't exist", er.NotFound, true, id)
	}
	return nil
}

func (c *Database) ListDocumentChunksForDocument(ctx context.Context, docID string) ([]aggregates.DocumentChunk, error) {
	tx, qtx, rollbackFn, err := c.beginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackFn()
	exists, err := c.documentExists(qtx, ctx, docID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, er.Newf("document %s doesn't exist", er.NotFound, true, docID)
	}
	chunks, err := qtx.ListDocumentChunksForDocument(ctx, pgxID(docID))
	if err != nil {
		return nil, err
	}
	result := []aggregates.DocumentChunk{}

	for _, chunk := range chunks {
		chunk := aggregates.DocumentChunk{
			ID:         chunk.ID.String(),
			CreatedAt:  chunk.CreatedAt.Time,
			Fragment:   chunk.Fragment.String,
			DocumentID: docID,
			Embedding:  chunk.Embedding.Slice(),
		}
		result = append(result, chunk)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}
