package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/appclacks/maizai/cmd"
	"github.com/appclacks/maizai/config"
	"github.com/appclacks/maizai/internal/database"
	mhttp "github.com/appclacks/maizai/internal/http"
	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/internal/http/handlers"
	aimock "github.com/appclacks/maizai/mocks/github.com/appclacks/maizai/pkg/rag"
	"github.com/appclacks/maizai/pkg/assistant"
	ct "github.com/appclacks/maizai/pkg/context"
	"github.com/appclacks/maizai/pkg/rag"
	ragdata "github.com/appclacks/maizai/pkg/rag/aggregates"
	"github.com/appclacks/maizai/pkg/shared"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testCase struct {
	name         string
	path         string
	pathFn       func() string
	body         string
	bodyFn       func() string
	method       string
	expectedBody string
	status       int
	callback     func(t *testing.T, response []byte) error
}

var listResponse client.ListContextOutput
var listDocumentsResponse client.ListDocumentsOutput
var contextResponse shared.Context
var ListchunksResponse client.ListDocumentChunksOutput

func sortMeta(i, j int) bool {
	res := strings.Compare(listResponse.Contexts[i].Name, listResponse.Contexts[j].Name)
	return res < 0
}

var cases = []testCase{
	{
		name:         "list context",
		path:         "/api/v1/context",
		method:       http.MethodGet,
		expectedBody: "[]",
		status:       200,
	},
	{
		name:         "created context",
		path:         "/api/v1/context",
		method:       http.MethodPost,
		body:         `{"name":"foo","description":"bar"}`,
		expectedBody: "context created",
		status:       200,
	},
	{
		name:         "list context after creation",
		path:         "/api/v1/context",
		method:       http.MethodGet,
		expectedBody: "foo",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &listResponse); err != nil {
				return err
			}
			assert.Len(t, listResponse.Contexts, 1)
			assert.Equal(t, listResponse.Contexts[0].Name, "foo")
			assert.Equal(t, listResponse.Contexts[0].Description, "bar")
			assert.NotZero(t, listResponse.Contexts[0].CreatedAt)
			return nil
		},
	},
	{
		name: "get context",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "foo",
		status:       200,
	},
	{
		name: "get context, not found",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", uuid.NewString())
		},
		method:       http.MethodGet,
		expectedBody: "doesn't exist",
		status:       404,
	},
	{
		name:   "create second context",
		path:   "/api/v1/context",
		method: http.MethodPost,
		bodyFn: func() string {
			return fmt.Sprintf(`{"name":"bar","description":"baz", "messages":[{"role":"user","content":"hello"},{"role":"user","content":"goodbye"}],"sources":{"contexts":["%s"]}}`, listResponse.Contexts[0].ID)
		},
		expectedBody: "context created",
		status:       200,
	},
	{
		name:         "list two contexts",
		path:         "/api/v1/context",
		method:       http.MethodGet,
		expectedBody: "foo",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &listResponse); err != nil {
				return err
			}
			sort.Slice(listResponse.Contexts, sortMeta)
			assert.Len(t, listResponse.Contexts, 2)
			assert.Equal(t, listResponse.Contexts[0].Name, "bar")
			assert.Equal(t, listResponse.Contexts[0].Description, "baz")
			assert.NotZero(t, listResponse.Contexts[0].CreatedAt)
			assert.Equal(t, listResponse.Contexts[1].Name, "foo")
			assert.Equal(t, listResponse.Contexts[1].Description, "bar")
			assert.NotZero(t, listResponse.Contexts[1].CreatedAt)
			return nil
		},
	},
	{
		name: "get first context",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Equal(t, contextResponse.Description, "baz")
			assert.Len(t, contextResponse.Messages, 2)
			assert.Equal(t, contextResponse.Messages[0].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[0].Content, "hello")
			assert.Equal(t, contextResponse.Messages[1].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[1].Content, "goodbye")
			assert.Equal(t, contextResponse.Sources.Contexts[0], listResponse.Contexts[1].ID)
			assert.NotZero(t, contextResponse.CreatedAt)
			return nil
		},
	},
	{
		name: "delete context source",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s/sources/context/%s", listResponse.Contexts[0].ID, listResponse.Contexts[0].Sources.Contexts[0])
		},
		method:       http.MethodDelete,
		expectedBody: "source deleted",
		status:       200,
	},
	{
		name: "get context after source deletion",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			contextResponse = shared.Context{}
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Len(t, contextResponse.Sources.Contexts, 0)
			return nil
		},
	},
	{
		name: "add context source",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s/sources/context/%s", listResponse.Contexts[0].ID, listResponse.Contexts[1].ID)
		},
		method:       http.MethodPost,
		expectedBody: "added",
		status:       200,
	},
	{
		name: "get context after source addition",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Len(t, contextResponse.Sources.Contexts, 1)
			assert.Equal(t, contextResponse.Sources.Contexts[0], listResponse.Contexts[1].ID)
			return nil
		},
	},
	{
		name: "add messages to context",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s/message", listResponse.Contexts[0].ID)
		},
		method:       http.MethodPost,
		expectedBody: "added",
		body:         `{"messages":[{"role":"user","content":"new1"},{"role":"assistant","content":"new2"}]}`,
		status:       200,
	},
	{
		name: "get context after messages addition",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Equal(t, contextResponse.Description, "baz")
			assert.Len(t, contextResponse.Messages, 4)
			assert.Equal(t, contextResponse.Messages[0].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[0].Content, "hello")
			assert.Equal(t, contextResponse.Messages[1].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[1].Content, "goodbye")
			assert.Equal(t, contextResponse.Messages[2].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[2].Content, "new1")
			assert.Equal(t, contextResponse.Messages[3].Role, shared.AssistantRole)
			assert.Equal(t, contextResponse.Messages[3].Content, "new2")
			assert.Equal(t, contextResponse.Sources.Contexts[0], listResponse.Contexts[1].ID)
			assert.NotZero(t, contextResponse.CreatedAt)
			return nil
		},
	},
	{
		name: "update context message",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/message/%s", contextResponse.Messages[1].ID)
		},
		method:       http.MethodPut,
		expectedBody: "message updated",
		body:         `{"role":"assistant","content":"updated"}`,
		status:       200,
	},
	{
		name: "get context after messages update",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Equal(t, contextResponse.Description, "baz")
			assert.Len(t, contextResponse.Messages, 4)
			assert.Equal(t, contextResponse.Messages[1].Role, shared.AssistantRole)
			assert.Equal(t, contextResponse.Messages[1].Content, "updated")
			return nil
		},
	},
	{
		name: "delete context message",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/message/%s", contextResponse.Messages[1].ID)
		},
		method:       http.MethodDelete,
		expectedBody: "message deleted",
		status:       200,
	},
	{
		name: "get context after messages addition",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodGet,
		expectedBody: "bar",
		status:       200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &contextResponse); err != nil {
				return err
			}
			assert.Equal(t, contextResponse.Name, "bar")
			assert.Equal(t, contextResponse.Description, "baz")
			assert.Len(t, contextResponse.Messages, 3)
			assert.Equal(t, contextResponse.Messages[0].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[0].Content, "hello")
			assert.Equal(t, contextResponse.Messages[1].Role, shared.UserRole)
			assert.Equal(t, contextResponse.Messages[1].Content, "new1")
			assert.Equal(t, contextResponse.Messages[2].Role, shared.AssistantRole)
			assert.Equal(t, contextResponse.Messages[2].Content, "new2")
			assert.Equal(t, contextResponse.Sources.Contexts[0], listResponse.Contexts[1].ID)
			assert.NotZero(t, contextResponse.CreatedAt)
			return nil
		},
	},
	{
		name:   "create document",
		path:   "/api/v1/document",
		method: http.MethodPost,
		bodyFn: func() string {
			return `{"name":"doc1","description":"desc1"}`
		},
		expectedBody: "document created",
		status:       200,
	},
	{
		name:   "list documents",
		path:   "/api/v1/document",
		method: http.MethodGet,
		status: 200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &listDocumentsResponse); err != nil {
				return err
			}
			assert.Len(t, listDocumentsResponse.Documents, 1)
			assert.NoError(t, uuid.Validate(listDocumentsResponse.Documents[0].ID))
			assert.Equal(t, listDocumentsResponse.Documents[0].Name, "doc1")
			assert.Equal(t, listDocumentsResponse.Documents[0].Description, "desc1")
			assert.NotZero(t, listDocumentsResponse.Documents[0].CreatedAt)
			return nil
		},
	},
	{
		name: "get document",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document/%s", listDocumentsResponse.Documents[0].ID)
		},
		method: http.MethodGet,
		status: 200,
		callback: func(t *testing.T, response []byte) error {
			var doc ragdata.Document
			t.Helper()
			if err := json.Unmarshal(response, &doc); err != nil {
				return err
			}
			assert.NoError(t, uuid.Validate(doc.ID))
			assert.Equal(t, doc.Name, "doc1")
			assert.Equal(t, doc.Description, "desc1")
			assert.NotZero(t, doc.CreatedAt)
			return nil
		},
	},
	{
		name: "Embed document",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document/%s", listDocumentsResponse.Documents[0].ID)
		},
		body:         `{"provider":"mistral","model":"mistral-embed","input":"trololo"}`,
		method:       http.MethodPost,
		expectedBody: "document chunk created",
		status:       200,
	},
	{
		name: "Embed document again",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document/%s", listDocumentsResponse.Documents[0].ID)
		},
		body:         `{"provider":"mistral","model":"mistral-embed","input":"trololo"}`,
		method:       http.MethodPost,
		expectedBody: "document chunk created",
		status:       200,
	},
	{
		name: "List document chunks",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document/%s/chunks", listDocumentsResponse.Documents[0].ID)
		},
		method: http.MethodGet,
		status: 200,
		callback: func(t *testing.T, response []byte) error {
			t.Helper()
			if err := json.Unmarshal(response, &ListchunksResponse); err != nil {
				return err
			}
			assert.Len(t, ListchunksResponse.Chunks, 2)
			assert.Equal(t, ListchunksResponse.Chunks[0].DocumentID, listDocumentsResponse.Documents[0].ID)
			assert.Equal(t, ListchunksResponse.Chunks[0].Fragment, "trololo")
			return nil
		},
	},
	{
		name: "delete document chunk",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document-chunk/%s", ListchunksResponse.Chunks[0].ID)
		},
		method:       http.MethodDelete,
		expectedBody: "document chunk deleted",
		status:       200,
	},
	{
		name: "delete document",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/document/%s", listDocumentsResponse.Documents[0].ID)
		},
		method:       http.MethodDelete,
		expectedBody: "document deleted",
		status:       200,
	},
	{
		name: "delete context 1",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[0].ID)
		},
		method:       http.MethodDelete,
		expectedBody: "context deleted",
		status:       200,
	},
	{
		name: "delete context 2",
		pathFn: func() string {
			return fmt.Sprintf("/api/v1/context/%s", listResponse.Contexts[1].ID)
		},
		method:       http.MethodDelete,
		expectedBody: "context deleted",
		status:       200,
	},
	{
		name:         "list contexts after deletion",
		path:         "/api/v1/context",
		method:       http.MethodGet,
		expectedBody: "[]",
		status:       200,
	},
}

func httpTest(t *testing.T, client *http.Client, c testCase) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	if c.bodyFn != nil {
		body = strings.NewReader(c.bodyFn())
	}
	path := c.path
	if c.pathFn != nil {
		path = c.pathFn()
	}
	req, err := http.NewRequestWithContext(ctx, c.method, fmt.Sprintf("http://localhost:3333%s", path), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != c.status {
		return fmt.Errorf("invalid status code (got %d)", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if c.expectedBody != "" {
		if !strings.Contains(string(bodyBytes), c.expectedBody) {
			return fmt.Errorf("invalid body (got %s, expected %s)", string(bodyBytes), c.expectedBody)
		}
	}
	if c.callback != nil {
		err := c.callback(t, bodyBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestIntegration(t *testing.T) {
	aiMock := aimock.NewMockAI(t)
	os.Setenv("MAIZAI_ANTHROPIC_API_KEY", "random_api_key")
	os.Setenv("MAIZAI_MISTRAL_API_KEY", "random_api_key")
	embedding := []float32{}
	for i := 0; i < 1024; i++ {
		embedding = append(embedding, float32(i))
	}
	aiMock.On("Embedding", mock.Anything, mock.Anything).Return(
		&ragdata.EmbeddingAnswer{
			InputTokens:  10,
			OutputTokens: 20,
			Data: []ragdata.Embedding{
				{
					Embedding: embedding,
				},
			},
		}, nil)

	registry := prometheus.NewRegistry()
	config, err := config.Load()
	assert.NoError(t, err)
	db, err := database.New(config.Store.PostgreSQL)
	assert.NoError(t, err)
	clients, err := cmd.BuildProviders(config.Providers)
	assert.NoError(t, err)
	manager := ct.New(db)
	assert.NoError(t, err)
	embeddingClients := map[string]rag.AI{}
	embeddingClients["mistral"] = aiMock

	rag := rag.New(db, embeddingClients)
	ai := assistant.New(clients, manager, rag)

	handlersBuilder := handlers.NewBuilder(ai, manager, rag)
	server, err := mhttp.New(config.HTTP, registry, handlersBuilder)
	assert.NoError(t, err)

	go func() {
		err := server.Start()
		assert.NoError(t, err)
	}()

	time.Sleep(2 * time.Second)

	assert.NoError(t, err)

	httpClient := http.DefaultClient
	for _, c := range cases {
		err := httpTest(t, httpClient, c)
		assert.NoError(t, err, c.name)
	}

	err = server.Stop()
	assert.NoError(t, err)
}
