# MaizAI, your AI agent toolbox

MaizAI is an API-first tool that abstracts AI providers and models (currently, Anthropic and Mistral are supported) and helps you manage contexts when interacting with them.

I built MaizAI to be able to easily build AI agents through simple HTTP calls, while having MaiZAI handling AI providers, models, contexts storage and retrieval, and RAG for me, and to be able to quickly iterate to find the best models and prompts for my use cases.

Key features:

- API-first (see OpenAPI spec)
- Contexts management: Create, update, delete contexts and use them when sending queries to AI providers. MaiZAI also supports dependencies between contexts to favor contexts reusing across messages.
- Streaming support (using SSE)
- RAG support using [pgvector](https://github.com/pgvector/pgvector), with an API to generate vectors from text: send text data and the embedding provider and model to use, and MaiZAI will take care of querying the embedding provider and store the result for later use
- Advanced observability: Prometheus metrics, Opentelemetry traces implementing [Gen AI semantic conventions for spans attributes](https://opentelemetry.io/docs/specs/semconv/gen-ai/) (input and output tokens, models used, provider...). The SQL layer (including the RAG one) and the HTTP server is also fully covered by tracing.
- Simple deployment: MaizAI requires only a PostgreSQL database to run and is distributed as a single binary (a Docker image is also available)
- Extensible: adding more AI or embedding providers is easy
- A CLI to easily interact with MaizAI API

## Supported AI providers

In order to use MaizAI, you need to configure at least one AI provider. If you want to use MaizAI RAG feature, you need at least one provider supporting embedding.

Feel free to open issues if you want to see a new provider being supported.

| Provider | Message | Embedding | Env variable
| --- | --- | --- | --- |
| Mistral | [ ] | [ ] | MAIZAI_MISTRAL_API_KEY |
| Anthropic | [ ] | [x] | MAIZAI_ANTHROPIC_API_KEY |

## Getting started

You need a [PostgreSQL](https://www.postgresql.org/) instance to run MaizAI.

### Configuration

Here is the list of available environment variables to configure MaizAI

| Env variable | Description | Default
| --- | --- | --- |
| MAIZAI_HTTP_HOST | MaizAI HTTP server host | 0.0.0.0 |
| MAIZAI_HTTP_PORT | MaizAI HTTP server port | 3333 |
| MAIZAI_HTTP_TLS_KEY_PATH | MaizAI HTTP server tls key path for mTLS |  |
| MAIZAI_HTTP_TLS_CERT_PATH | MaizAI HTTP server tls cert path for mTLS |  |
| MAIZAI_HTTP_TLS_CACERT_PATH | MaizAI HTTP server tls cacert path for mTLS |  |
| MAIZAI_HTTP_TLS_INSECURE | MaizAI HTTP server tls insecure | false |
| MAIZAI_HTTP_TLS_SERVER_NAME | MaizAI HTTP server tls server name (sni) |  |
| MAIZAI_POSTGRESQL_USERNAME | MaizAI PostgreSQL database username |  |
| MAIZAI_POSTGRESQL_PASSWORD | MaizAI PostgreSQL database password |  |
| MAIZAI_POSTGRESQL_DATABASE | MaizAI PostgreSQL database name |  |
| MAIZAI_POSTGRESQL_HOST | MaizAI PostgreSQL database host |  |
| MAIZAI_POSTGRESQL_PORT | MaizAI PostgreSQL database port |  |
| MAIZAI_POSTGRESQL_SSL_MODE | MaizAI PostgreSQL ssl mode |  |

OpenTelemetry traces can be optionally configured using the [standard Otel environment variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).

### Using Docker Compose

You can use this Docker Compose file to run PostgreSQL and Jaeger (Opentelemetry backend):

```bash
docker compose up -d
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces
// You need at least one AI provider configured to run MaizAI
// export MAIZAI_ANTHROPIC_API_KEY=api_key
// export MAIZAI__MISTRAL_API_KEY=api_key
maizai server
// MaizAI is now running on port 3333
```

### Using MaizAI's CLI to interact with the server

This section explains the main MaizAI's CLI commands. You can use the `--help` flag at every layer to get information about the available commands and options, for example: `maizai --help`, `maizai conversation --help`

#### Conversation

Let's start a conversation using the Mistral provider and its mistral-small-latest model:

```
maizai conversation \
  --provider mistral \
  --model mistral-small-latest \
  --system "you're a general purpose AI assistant" \
  --interactive \
  --context-name "my-context"
Hello, I'm your AI assistant. Ask me anything:

Why is the sky blue?

Answer (input tokens 18, output tokens 162):

The sky appears blue due to a particular type of scattering called Rayleigh scattering. As light from the sun reaches Earth's atmosphere, it is scattered in different directions by the gas molecules and tiny particles in the air. Blue light is scattered more than other colors because it travels in shorter, smaller waves. This is why we perceive the sky as blue most of the time. [...]

Anything else (write 'exit' to exit the program)?
```

This command will create a new context :

```
maizai context list

{
  "contexts": [
    {
      "id": "01eff873-1e30-65de-8980-a6567a017827",
      "name": "my-context",
      "created-at": "2025-03-03T21:04:38.320688Z",
      "sources": {}
    }
  ]
}
```

MaizAI automatically created a context when interacting with the AI provider. You can also choose to use an existing context by passing a `--context-id` flag to the `maizai conversation` command instead of a context name. This also allows you to start a conversation when you stopped it at any time.

You can also run non-interactive conversations by removing the `--interactive` flag and by passing your prompt using `--prompt`:

```
maizai conversation --provider mistral --model mistral-small-latest --system "you're a general purpose AI assistant" --context-name "my-context" --prompt "Why is the sky blue?"

```json
{
  "result": [
    {
      "text": "The sky appears blue due to a particular type of scattering called Rayleigh scattering. [...]"
    }
  ],
  "input-tokens": 18,
  "output-tokens": 249,
  "context": "01eff876-89f7-623e-8980-a6567a017827"
}
```

#### Managing contexts

Contexts are store inside PostgreSQL. Messages (inputs and outputs) are appened to the context and provided to the AI provider for each message sent to it.
You can get a context by ID:

```
maizai context get --id 01eff873-1e30-65de-8980-a6567a017827 | jq
{
  "id": "01eff873-1e30-65de-8980-a6567a017827",
  "name": "my-context",
  "sources": {},
  "messages": [
    {
      "id": "01eff873-1e30-6b20-8980-a6567a017827",
      "role": "user",
      "content": "Why is the sky blue?\n",
      "created-at": "2025-03-03T21:04:38.320683Z"
    },
    {
      "id": "01eff873-201d-6c7b-8980-a6567a017827",
      "role": "assistant",
      "content": "The sky appears blue due to a particular type of scattering called Rayleigh scattering. As the sun's light reaches Earth's atmosphere,[...]",
      "created-at": "2025-03-03T21:04:41.553832Z"
    }
  ],
  "created-at": "2025-03-03T21:04:38.320688Z"
}
```


You can also add, delete, or update messages using the `maizai context message add/delete/update` subcommands. This allows you to build a your own context without having to interact with an AI provider:

For example, you can add a new message to an existing context:

```
maizai context message add --id 01eff873-1e30-65de-8980-a6567a017827 --message "user:Is Mars sky blue?"
{"messages":["messages added to context"]}
```

**Contexts sources**

Sometimes, you'll want to reuse a context but without appending new messages to it. For example, you may want to have some generic-purpose contexts shared across all your conversations.

You can do that on MaizAI by using contexts sources. Let's create a new context:

```
maizai context create --name "my-source" --message "user:This is the message that will be sent to the AI provider"
{"messages":["context created"]}
```

Now, retrieving the new context ID by running `maizai context list` and add it as a source to our existing context:

```
maizai context source add-context --id 01eff873-1e30-65de-8980-a6567a017827 --source-context-id 01eff875-4638-6fa8-8980-a6567a017827
{"messages":["Context source added"]}
```

If you list contexts again, you'll see that our first context now has a source:

```json
{
  "id": "01eff873-1e30-65de-8980-a6567a017827",
  "name": "my-context",
  "created-at": "2025-03-03T21:04:38.320688Z",
  "sources": {
    "contexts": [
      "01eff875-4638-6fa8-8980-a6567a017827"
    ]
  }
}
```

You can remove a source to an existing context by using `maizai context source remove-context`. You can also add sources to newly created contexts (on `maizai context` and `maizai conversation`) by specifying the `--source-context` flag.

#### Using RAG

To use MaizAI's rag feature, you need a Mistral AI account and an API key (`MAIZAI_MISTRAL_API_KEY` env variable). Then, create a document. Documents are used in MaizAI to group chunks coming from the same source (an article, a book...) together:

```
maizai document create --name "my-doc" --description "description"
{"messages":["document created"]}
```

You can use the `document get`, `document delete`, `document list` subcommands to manage documents.

Them, embed content for this document:

```
maizai document embed
  --document-id 01eff9c9-a727-6b94-8dc4-a6567a017827
  --input "Mathieu Corbin is the author of the mcorbin.fr blog"
{"messages":["document chunk created"]}
```

In the background, MaizAI will:

- Reach out to the configured AI provider embedding API to convert the input to a vector
- Store the result in MaizAI database

The CLI (and MaizAI API) allows you to manage chunks: listing chunks for a document using `maizai document list-chunks`.

You can now query the rag using the conversation API. In this example, we ask the RAG information about Mathieu Corbin, and limit the number of chunks returned to 1. The data retrieved will replace the `{maizai_rag_data}` placeholder in the prompt.

```
maizai conversation
  --provider mistral
  --model mistral-small-latest
  --system "you're a general purpose AI assistant"
  --context-name "context-with-rag"
  --rag-model mistral-embed
  --rag-provider mistral
  --rag-limit 1
  --rag-input "Information about Mathieu Corbin"
  --prompt "Who is Mathieu Corbin? Use this context to help you: {maizai_rag_data}"

{
  "result": [
    {
      "text": "Based on the context provided, Mathieu Corbin is the author of the mcorbin.fr blog. However, without additional information, I can't provide more details about his background, expertise, or the content of his blog. If you have more context or specific questions about Mathieu Corbin or his blog, feel free to share!"
    }
  ],
  "input-tokens": 38,
  "output-tokens": 66,
  "context": "01eff9cb-2fa3-6bae-8dc4-a6567a017827"
}
```

You can also query MaizAI's RAG by using the `embedding match` command. This can be helpful to validate that your RAG is returning proper information:

```
maizai embedding match --input "Information about Mathieu Corbin" --limit 1 --model mistral-embed --provider mistral | jq
{
  "chunks": [
    {
      "id": "01eff9ca-bc61-694f-8dc4-a6567a017827",
      "document-id": "01eff9c9-a727-6b94-8dc4-a6567a017827",
      "fragment": "Mathieu Corbin is the author of the mcorbin.fr blog",
      "created-at": "2025-03-05T14:04:21.096893Z"
    }
  ]
}
```
