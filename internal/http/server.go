package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/appclacks/maizai/internal/http/client"
	"github.com/appclacks/maizai/internal/http/handlers"
	"github.com/appclacks/maizai/internal/tls"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type Configuration struct {
	Host       string `env:"MAIZAY_HTTP_HOST, default=0.0.0.0"`
	Port       uint32 `env:"MAIZAY_HTTP_PORT, default=3333"`
	Key        string `env:"MAIZAY_HTTP_TLS_KEY_PATH"`
	Cert       string `env:"MAIZAY_HTTP_TLS_CERT_PATH"`
	Cacert     string `env:"MAIZAY_HTTP_TLS_CACERT_PATH"`
	Insecure   bool   `env:"MAIZAY_HTTP_TLS_INSECURE"`
	ServerName string `env:"MAIZAY_HTTP_TLS_SERVER_NAME"`
}

type Server struct {
	config *Configuration
	e      *echo.Echo
	wg     sync.WaitGroup
}

func New(config Configuration, registry *prometheus.Registry, builder *handlers.Builder) (*Server, error) {
	if config.Host == "" || config.Port == 0 {
		return nil, errors.New("Invalid HTTP configuration: host and port are mandatory")
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	buckets := []float64{
		0.05, 0.1, 0.2, 0.4, 0.8, 1,
		1.5, 2, 3, 5}

	respCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Count the number of HTTP responses",
		},
		[]string{"method", "status", "path"})

	err := registry.Register(respCounter)
	if err != nil {
		return nil, err
	}

	reqHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_duration_second",
			Help:    "Time to execute http requests",
			Buckets: buckets,
		},
		[]string{"method", "path"})

	err = registry.Register(reqHistogram)
	if err != nil {
		return nil, err
	}
	e.HTTPErrorHandler = handlers.ErrorHandler()

	e.Use(otelecho.Middleware("maizai"))
	e.Use(metricsMiddleware(reqHistogram, respCounter))
	e.GET("/healthz", func(ec echo.Context) error {
		return ec.JSON(http.StatusOK, "ok")
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	definitions := []apiv1{
		{
			path:        "/conversation",
			method:      http.MethodPost,
			handler:     builder.Conversation,
			payload:     client.CreateConversationInput{},
			response:    client.ConversationAnswer{},
			description: "Send a message to the AI provider. If a context ID is passed as parameter, use this context as a base. Else, a new context whose name will be the context named as parameter will be created.",
		},
		{
			path:        "/context",
			method:      http.MethodGet,
			handler:     builder.ListContexts,
			payload:     nil,
			response:    client.ListContextOutput{},
			description: "List contexts",
		},
		{
			path:        "/context/:id",
			method:      http.MethodGet,
			handler:     builder.GetContext,
			payload:     client.GetContextInput{},
			response:    client.Context{},
			description: "Get a context by ID",
		},
		{
			path:        "/context",
			method:      http.MethodPost,
			handler:     builder.CreateContext,
			payload:     client.CreateContextInput{},
			response:    client.Response{},
			description: "Create a new context",
		},
		{
			path:        "/context/:id",
			method:      http.MethodDelete,
			handler:     builder.DeleteContext,
			payload:     client.DeleteContextInput{},
			response:    client.Response{},
			description: "Delete a context by ID",
		},
		{
			path:        "/context/:id/sources/context/:source-context-id",
			method:      http.MethodDelete,
			handler:     builder.DeleteContextSourceContext,
			payload:     client.DeleteContextSourceContextInput{},
			response:    client.Response{},
			description: "Remove a context used as a source for a given context",
		},
		{
			path:        "/context/:id/sources/context/:source-context-id",
			method:      http.MethodPost,
			handler:     builder.CreateContextSourceContext,
			payload:     client.CreateContextSourceContextInput{},
			response:    client.Response{},
			description: "Add a context as a source for a given context",
		},
		{
			path:        "/context/:id/message",
			method:      http.MethodPost,
			handler:     builder.AddMessagesToContext,
			payload:     client.AddMessagesToContextInput{},
			response:    client.Response{},
			description: "Add new messages for a given context",
		},
		{
			path:        "/context/:id/message",
			method:      http.MethodDelete,
			handler:     builder.DeleteContextMessages,
			payload:     client.DeleteContextMessagesInput{},
			response:    client.Response{},
			description: "Delete all messages for a given context",
		},
		{
			path:        "/message/:id",
			method:      http.MethodPut,
			handler:     builder.UpdateContextMessage,
			payload:     client.UpdateContextMessageInput{},
			response:    client.Response{},
			description: "Update an existing message",
		},
		{
			path:        "/message/:id",
			method:      http.MethodDelete,
			handler:     builder.DeleteContextMessage,
			payload:     client.DeleteContextMessageInput{},
			response:    client.Response{},
			description: "Delete a message by ID",
		},
		{
			path:        "/document",
			method:      http.MethodGet,
			handler:     builder.ListDocuments,
			payload:     nil,
			response:    client.ListDocumentsOutput{},
			description: "List documents",
		},
		{
			path:        "/document",
			method:      http.MethodPost,
			handler:     builder.CreateDocument,
			payload:     client.CreateDocumentInput{},
			response:    client.Response{},
			description: "Create a new document",
		},
		{
			path:        "/document/:document-id",
			method:      http.MethodPost,
			handler:     builder.EmbedDocument,
			payload:     client.EmbedDocumentInput{},
			response:    client.Response{},
			description: "Embed the input passed as parameter for the given document, to use it later in MaizAI's RAG",
		},
		{
			path:        "/document/:id",
			method:      http.MethodGet,
			handler:     builder.GetDocument,
			payload:     client.GetDocumentInput{},
			response:    client.Document{},
			description: "Get a document by ID",
		},
		{
			path:        "/document/:id/chunks",
			method:      http.MethodGet,
			handler:     builder.ListDocumentChunksForDocument,
			payload:     client.ListDocumentChunksForDocumentInput{},
			response:    client.ListDocumentChunksOutput{},
			description: "List chunks for a given document",
		},
		{
			path:        "/document/:id",
			method:      http.MethodDelete,
			handler:     builder.DeleteDocument,
			payload:     client.DeleteDocumentInput{},
			response:    client.Response{},
			description: "Delete a document by ID",
		},
		{
			path:        "/document-chunk/:id",
			method:      http.MethodDelete,
			handler:     builder.DeleteDocumentChunk,
			payload:     client.DeleteDocumentChunkInput{},
			response:    client.Response{},
			description: "Delete a document chunk by ID",
		},
		{
			path:        "/document-chunk",
			method:      http.MethodPut,
			handler:     builder.MatchChunk,
			payload:     client.RagSearchQuery{},
			response:    client.ListDocumentChunksOutput{},
			description: "Return chunks matching the provided input",
		},
	}

	err = openapiSpec(e, definitions)
	if err != nil {
		return nil, err
	}
	//apiGroup := e.Group("/api/v1")
	//apiGroup.POST("/conversation", builder.Conversation)
	//apiGroup.GET("/context", builder.ListContexts)
	//apiGroup.GET("/context/:id", builder.GetContext)
	//apiGroup.POST("/context", builder.CreateContext)
	//apiGroup.DELETE("/context/:id", builder.DeleteContext)
	//apiGroup.DELETE("/context/:id/sources/context/:source-context-id", builder.DeleteContextSourceContext)
	//apiGroup.POST("/context/:id/sources/context/:source-context-id", builder.CreateContextSourceContext)
	//apiGroup.POST("/context/:id/message", builder.AddMessagesToContext)
	//apiGroup.PUT("/message/:id", builder.UpdateContextMessage)
	//apiGroup.DELETE("/message/:id", builder.DeleteContextMessage)
	//apiGroup.GET("/document", builder.ListDocuments)
	//apiGroup.POST("/document", builder.CreateDocument)
	//apiGroup.POST("/document/:document-id", builder.EmbedDocument)
	//apiGroup.GET("/document/:id", builder.GetDocument)
	//apiGroup.GET("/document/:id/chunks", builder.ListDocumentChunksForDocument)
	//apiGroup.DELETE("/document/:id", builder.DeleteDocument)
	//apiGroup.DELETE("/document-chunk/:id", builder.DeleteDocumentChunk)
	//apiGroup.PUT("/document-chunk", builder.MatchChunk)

	return &Server{
		config: &config,
		e:      e,
	}, nil
}

func (s *Server) Start() error {
	address := fmt.Sprintf("[%s]:%d", s.config.Host, s.config.Port)
	slog.Info(fmt.Sprintf("http server starting on %s", address))
	if s.config.Cert != "" {
		slog.Info("tls is enabled on the http server")
		tlsConfig, err := tls.GetTLSConfig(s.config.Key, s.config.Cert, s.config.Cacert, s.config.ServerName, s.config.Insecure)
		if err != nil {
			return err
		}

		s.e.TLSServer.TLSConfig = tlsConfig
		tlsServer := s.e.TLSServer
		tlsServer.Addr = address
		if !s.e.DisableHTTP2 {
			tlsServer.TLSConfig.NextProtos = append(tlsServer.TLSConfig.NextProtos, "h2")
		}
	}

	go func() {
		defer s.wg.Done()
		var err error
		if s.config.Cert != "" {
			err = s.e.StartServer(s.e.TLSServer)
		} else {
			err = s.e.Start(address)

		}
		if err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("http server error: %s", err.Error()))
			os.Exit(2)
		}

	}()
	s.wg.Add(1)
	return nil
}

func (s *Server) Stop() error {
	slog.Info("stopping the http server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.e.Shutdown(ctx)
	s.wg.Wait()
	if err != nil {
		return err
	}
	slog.Info("http server stopped")
	return nil
}
