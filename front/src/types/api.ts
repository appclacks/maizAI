export interface ClientNewMessage {
  role: string;
  content: string;
}

export interface ClientQueryOptions {
  model: string;
  'max-tokens': number;
  provider: string;
  temperature?: number;
  system?: string;
  rag?: ClientRagSearchQuery;
}

export interface ClientRagSearchQuery {
  input: string;
  limit: number;
  model: string;
  provider: string;
}

export interface ClientContextOptions {
  name: string;
  description?: string;
  sources?: ClientContextSources;
}

export interface ClientContextSources {
  contexts?: string[];
}

export interface ClientCreateConversationInput {
  messages: ClientNewMessage[];
  'context-id'?: string;
  'new-context'?: ClientContextOptions;
  'query-options'?: ClientQueryOptions;
  stream?: boolean;
  'system-prompt-id'?: string;
}

export interface ClientResult {
  text: string;
}

export interface ClientConversationAnswer {
  context: string;
  'input-tokens': number;
  'output-tokens': number;
  result: ClientResult[] | null;
}

// Context management types
export interface ClientMessage {
  id: string;
  role: string;
  content: string;
  'created-at': string;
}

export interface ClientContext {
  id: string;
  name: string;
  description?: string;
  'created-at': string;
  messages?: ClientMessage[];
  sources?: ClientContextSources;
}

export interface ClientContextMetadata {
  id: string;
  name: string;
  description?: string;
  'created-at': string;
  sources?: ClientContextSources;
}

export interface ClientListContextOutput {
  contexts: ClientContextMetadata[] | null;
}

export interface ClientCreateContextInput {
  name: string;
  description?: string;
  messages?: ClientNewMessage[] | null;
  sources?: ClientContextSources;
}

export interface ClientResponse {
  messages?: string[] | null;
}

export interface ClientUpdateContextMessageInput {
  role: string;
  content: string;
}

export interface ClientAddMessagesToContextInput {
  messages: ClientNewMessage[] | null;
}

// System Prompt management types
export interface ClientSystemPrompt {
  id: string;
  name: string;
  description?: string;
  content: string;
  'created-at': string;
}

export interface ClientListSystemPromptsOutput {
  system_prompts: ClientSystemPrompt[] | null;
}

export interface ClientCreateSystemPromptInput {
  name: string;
  description?: string;
  content: string;
}

export interface ClientUpdateSystemPromptInput {
  content: string;
}

// Document management types
export interface ClientDocument {
  id: string;
  name: string;
  description?: string;
  'created-at': string;
}

export interface ClientListDocumentsOutput {
  documents: ClientDocument[] | null;
}

export interface ClientCreateDocumentInput {
  name: string;
  description?: string;
}

export interface ClientDocumentChunk {
  id: string;
  'document-id': string;
  fragment: string;
  'created-at': string;
}

export interface ClientListDocumentChunksOutput {
  chunks: ClientDocumentChunk[] | null;
}

export interface ClientEmbedDocumentInput {
  model: string;
  input: string;
  provider: string;
}