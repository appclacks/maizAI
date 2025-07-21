'use client';

import { useState, useEffect, useCallback } from 'react';
import { 
  ClientDocument, 
  ClientListDocumentsOutput, 
  ClientCreateDocumentInput, 
  ClientEmbedDocumentInput,
  ClientDocumentChunk,
  ClientListDocumentChunksOutput,
  ClientResponse 
} from '@/types/api';

interface ApiError {
  message: string;
  status?: number;
  details?: string;
  timestamp: string;
}

// Chunking utility functions
const chunkText = (text: string, strategy: string, chunkSize: number, overlap: number): string[] => {
  if (strategy === 'none') {
    return [text];
  }

  switch (strategy) {
    case 'sentence':
      return chunkBySentences(text, chunkSize);
    case 'paragraph':
      return chunkByParagraphs(text);
    case 'fixed-overlap':
      return chunkByFixedSizeWithOverlap(text, chunkSize, overlap);
    case 'smart':
      return chunkSmartly(text, chunkSize, overlap);
    default:
      return [text];
  }
};

const chunkBySentences = (text: string, maxChunkSize: number): string[] => {
  const sentences = text.match(/[^\.!?]+[\.!?]+/g) || [text];
  const chunks: string[] = [];
  let currentChunk = '';

  for (const sentence of sentences) {
    const trimmedSentence = sentence.trim();
    if (!trimmedSentence) continue;

    if (currentChunk.length + trimmedSentence.length > maxChunkSize && currentChunk.length > 0) {
      chunks.push(currentChunk.trim());
      currentChunk = trimmedSentence;
    } else {
      currentChunk += (currentChunk ? ' ' : '') + trimmedSentence;
    }
  }

  if (currentChunk.trim()) {
    chunks.push(currentChunk.trim());
  }

  return chunks.filter(chunk => chunk.length > 0);
};

const chunkByParagraphs = (text: string): string[] => {
  return text
    .split(/\n\s*\n/)
    .map(para => para.trim())
    .filter(para => para.length > 0);
};

const chunkByFixedSizeWithOverlap = (text: string, chunkSize: number, overlap: number): string[] => {
  // Safety checks to prevent infinite loops
  if (chunkSize <= 0) return [text];
  if (overlap < 0) overlap = 0;
  if (overlap >= chunkSize) overlap = Math.floor(chunkSize * 0.5); // Max 50% overlap
  
  const chunks: string[] = [];
  let start = 0;
  const minAdvancement = Math.max(1, chunkSize - overlap); // Ensure we always advance

  while (start < text.length) {
    const end = Math.min(start + chunkSize, text.length);
    const chunk = text.slice(start, end).trim();
    
    if (chunk.length > 0) {
      chunks.push(chunk);
    }
    
    if (end >= text.length) break;
    
    // Ensure we always advance by at least minAdvancement
    const nextStart = start + minAdvancement;
    start = nextStart;
    
    // Safety break to prevent infinite loops (max 10000 chunks)
    if (chunks.length > 10000) {
      console.warn('Too many chunks generated, stopping to prevent browser crash');
      break;
    }
  }

  return chunks.filter(chunk => chunk.length > 0);
};

const chunkSmartly = (text: string, maxChunkSize: number, overlap: number): string[] => {
  // Safety checks
  if (maxChunkSize <= 0) return [text];
  if (overlap < 0) overlap = 0;
  
  const sentences = text.match(/[^\.!?]+[\.!?]+/g) || [text];
  const chunks: string[] = [];
  let currentChunk = '';

  for (let i = 0; i < sentences.length; i++) {
    const sentence = sentences[i].trim();
    if (!sentence) continue;

    const wouldExceedSize = currentChunk.length + sentence.length > maxChunkSize;
    
    if (wouldExceedSize && currentChunk.length > 0) {
      chunks.push(currentChunk.trim());
      
      // Create overlap by including some previous sentences
      const overlapSentences = Math.max(0, Math.min(3, Math.floor(overlap / 200))); // Max 3 sentences for overlap
      const startIndex = Math.max(0, i - overlapSentences);
      currentChunk = sentences.slice(startIndex, i + 1).join(' ').trim();
    } else {
      currentChunk += (currentChunk ? ' ' : '') + sentence;
    }
    
    // Safety break to prevent infinite loops
    if (chunks.length > 10000) {
      console.warn('Too many chunks generated, stopping to prevent browser crash');
      break;
    }
  }

  if (currentChunk.trim()) {
    chunks.push(currentChunk.trim());
  }

  return chunks.filter(chunk => chunk.length > 0);
};

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<ClientDocument[]>([]);
  const [documentChunks, setDocumentChunks] = useState<ClientDocumentChunk[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<ApiError | null>(null);
  const [selectedDocument, setSelectedDocument] = useState<ClientDocument | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [showEmbedForm, setShowEmbedForm] = useState(false);
  const [apiBaseUrl, setApiBaseUrl] = useState(process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080');

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });

  // Embed form state
  const [embedData, setEmbedData] = useState({
    input: '',
    model: 'mistral-embed',
    provider: 'mistral',
    chunkingStrategy: 'none',
    chunkSize: 1000,
    chunkOverlap: 200,
  });

  const [previewChunks, setPreviewChunks] = useState<string[]>([]);
  const [showChunkPreview, setShowChunkPreview] = useState(false);

  const loadDocuments = useCallback(async () => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/document`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListDocumentsOutput = await response.json();
      setDocuments(data.documents || []);
    } catch (error) {
      console.error('Error loading documents:', error);
      const apiError: ApiError = {
        message: 'Failed to load documents',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl]);

  const loadDocumentDetails = async (documentId: string) => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/document/${documentId}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientDocument = await response.json();
      setSelectedDocument(data);

      // Load document chunks
      const chunksResponse = await fetch(`${apiBaseUrl}/api/v1/document/${documentId}/chunks`);
      
      if (chunksResponse.ok) {
        const chunksData: ClientListDocumentChunksOutput = await chunksResponse.json();
        setDocumentChunks(chunksData.chunks || []);
      }
    } catch (error) {
      console.error('Error loading document details:', error);
      const apiError: ApiError = {
        message: 'Failed to load document details',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const createDocument = async () => {
    if (!formData.name.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const documentInput: ClientCreateDocumentInput = {
        name: formData.name,
        description: formData.description || undefined,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/document`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(documentInput),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to create document');
      }

      // Reset form and reload documents
      setFormData({ name: '', description: '' });
      setShowCreateForm(false);
      await loadDocuments();
    } catch (error) {
      console.error('Error creating document:', error);
      const apiError: ApiError = {
        message: 'Failed to create document',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const embedDocument = async () => {
    if (!selectedDocument || !embedData.input.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      // Generate chunks based on selected strategy
      const chunks = chunkText(
        embedData.input,
        embedData.chunkingStrategy,
        embedData.chunkSize,
        embedData.chunkOverlap
      );

      console.log(`Embedding ${chunks.length} chunk(s) for document ${selectedDocument.name}`);
      
      // Send each chunk as a separate embedding request
      const embedPromises = chunks.map(async (chunk, index) => {
        const embedInput: ClientEmbedDocumentInput = {
          model: embedData.model,
          input: chunk,
          provider: embedData.provider,
        };

        const response = await fetch(`${apiBaseUrl}/api/v1/document/${selectedDocument.id}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(embedInput),
        });

        if (!response.ok) {
          const errorData: ClientResponse = await response.json();
          throw new Error(`Failed to embed chunk ${index + 1}: ${errorData.messages?.join(', ') || response.statusText}`);
        }

        return response.json();
      });

      // Wait for all chunks to be embedded
      await Promise.all(embedPromises);

      // Reset embed form and reload document chunks
      setEmbedData({ 
        input: '', 
        model: 'mistral-embed', 
        provider: 'mistral',
        chunkingStrategy: 'none',
        chunkSize: 1000,
        chunkOverlap: 200,
      });
      setPreviewChunks([]);
      setShowChunkPreview(false);
      setShowEmbedForm(false);
      
      if (selectedDocument) {
        await loadDocumentDetails(selectedDocument.id);
      }
    } catch (error) {
      console.error('Error embedding document:', error);
      const apiError: ApiError = {
        message: 'Failed to embed document content',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const previewChunking = () => {
    if (!embedData.input.trim()) return;
    
    const chunks = chunkText(
      embedData.input,
      embedData.chunkingStrategy,
      embedData.chunkSize,
      embedData.chunkOverlap
    );
    
    setPreviewChunks(chunks);
    setShowChunkPreview(true);
  };

  const deleteDocument = async (documentId: string) => {
    if (!confirm('Are you sure you want to delete this document? This action cannot be undone.')) {
      return;
    }

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/document/${documentId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete document');
      }

      // If the deleted document was selected, clear selection
      if (selectedDocument?.id === documentId) {
        setSelectedDocument(null);
        setDocumentChunks([]);
      }

      // Reload documents
      await loadDocuments();
    } catch (error) {
      console.error('Error deleting document:', error);
      const apiError: ApiError = {
        message: 'Failed to delete document',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteDocumentChunk = async (chunkId: string) => {
    if (!confirm('Are you sure you want to delete this document chunk?')) {
      return;
    }

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/document-chunk/${chunkId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete document chunk');
      }

      // Reload document chunks
      if (selectedDocument) {
        await loadDocumentDetails(selectedDocument.id);
      }
    } catch (error) {
      console.error('Error deleting document chunk:', error);
      const apiError: ApiError = {
        message: 'Failed to delete document chunk',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const dismissError = () => {
    setApiError(null);
  };

  const cancelCreate = () => {
    setShowCreateForm(false);
    setFormData({ name: '', description: '' });
  };

  const cancelEmbed = () => {
    setShowEmbedForm(false);
    setEmbedData({ 
      input: '', 
      model: 'mistral-embed', 
      provider: 'mistral',
      chunkingStrategy: 'none',
      chunkSize: 1000,
      chunkOverlap: 200,
    });
    setPreviewChunks([]);
    setShowChunkPreview(false);
  };

  // Load documents on component mount
  useEffect(() => {
    loadDocuments();
  }, [loadDocuments]);

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Document Management</h1>
        <div className="flex space-x-4">
          <input
            type="url"
            value={apiBaseUrl}
            onChange={(e) => setApiBaseUrl(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="API Base URL"
          />
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
          >
            {showCreateForm ? 'Cancel' : 'Create Document'}
          </button>
        </div>
      </div>

      {/* Error Display */}
      {apiError && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3 flex-1">
              <h3 className="text-sm font-medium text-red-800">API Error</h3>
              <div className="mt-2 text-sm text-red-700">
                <p><strong>Message:</strong> {apiError.message}</p>
                {apiError.details && (
                  <p><strong>Details:</strong> {apiError.details}</p>
                )}
              </div>
              <button
                onClick={dismissError}
                className="mt-2 bg-red-100 text-red-800 px-3 py-1 rounded-md text-sm hover:bg-red-200"
              >
                Dismiss
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Form */}
      {showCreateForm && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Create New Document</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Name *
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter document name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                rows={3}
                placeholder="Enter document description (optional)"
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={createDocument}
                disabled={!formData.name.trim() || isLoading}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Creating...' : 'Create Document'}
              </button>
              <button
                onClick={cancelCreate}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Embed Form */}
      {showEmbedForm && selectedDocument && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Embed Content for: {selectedDocument.name}</h2>
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Content to Embed *
              </label>
              <textarea
                value={embedData.input}
                onChange={(e) => setEmbedData(prev => ({ ...prev, input: e.target.value }))}
                className="w-full p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                rows={8}
                placeholder="Enter the text content you want to embed for this document. For large texts, consider using chunking options below."
              />
              <p className="text-xs text-gray-500 mt-1">
                Character count: {embedData.input.length}
              </p>
            </div>

            {/* Chunking Configuration */}
            <div className="bg-gray-50 p-4 rounded-lg border">
              <h3 className="text-lg font-medium text-gray-900 mb-3">Text Chunking Options</h3>
              <p className="text-sm text-gray-600 mb-4">
                For large texts, chunking helps improve RAG performance by creating smaller, focused segments.
              </p>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Chunking Strategy
                  </label>
                  <select
                    value={embedData.chunkingStrategy}
                    onChange={(e) => setEmbedData(prev => ({ ...prev, chunkingStrategy: e.target.value }))}
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="none">No chunking (embed as single piece)</option>
                    <option value="sentence">Sentence-based chunking (recommended)</option>
                    <option value="paragraph">Paragraph-based chunking</option>
                    <option value="fixed-overlap">Fixed-size with overlap</option>
                    <option value="smart">Smart chunking (sentence + size limits)</option>
                  </select>
                </div>

                {embedData.chunkingStrategy !== 'none' && embedData.chunkingStrategy !== 'paragraph' && (
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Max Chunk Size (characters)
                      </label>
                      <input
                        type="number"
                        min="200"
                        max="4000"
                        step="100"
                        value={embedData.chunkSize}
                        onChange={(e) => setEmbedData(prev => ({ ...prev, chunkSize: parseInt(e.target.value) || 1000 }))}
                        className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                    </div>
                    {(embedData.chunkingStrategy === 'fixed-overlap' || embedData.chunkingStrategy === 'smart') && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Overlap Size (characters)
                        </label>
                        <input
                          type="number"
                          min="0"
                          max={Math.floor(embedData.chunkSize * 0.5)}
                          step="50"
                          value={embedData.chunkOverlap}
                          onChange={(e) => setEmbedData(prev => ({ ...prev, chunkOverlap: parseInt(e.target.value) || 0 }))}
                          className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                      </div>
                    )}
                  </div>
                )}

                {embedData.chunkingStrategy !== 'none' && embedData.input.trim() && (
                  <div className="flex space-x-2">
                    <button
                      type="button"
                      onClick={previewChunking}
                      className="px-4 py-2 bg-gray-500 text-white rounded-md hover:bg-gray-600 text-sm"
                    >
                      Preview Chunks
                    </button>
                    <div className="text-sm text-gray-500 flex items-center">
                      Will create ~{chunkText(embedData.input, embedData.chunkingStrategy, embedData.chunkSize, embedData.chunkOverlap).length} chunks
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Chunk Preview */}
            {showChunkPreview && previewChunks.length > 0 && (
              <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-medium text-gray-900">Chunk Preview ({previewChunks.length} chunks)</h4>
                  <button
                    onClick={() => setShowChunkPreview(false)}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div className="space-y-2 max-h-60 overflow-y-auto">
                  {previewChunks.map((chunk, index) => (
                    <div key={index} className="bg-white p-3 rounded border text-sm">
                      <div className="flex justify-between items-center mb-2">
                        <span className="text-xs font-medium text-blue-600">Chunk {index + 1}</span>
                        <span className="text-xs text-gray-500">{chunk.length} chars</span>
                      </div>
                      <div className="text-gray-700 max-h-24 overflow-y-auto border border-gray-100 rounded p-2 bg-gray-50">
                        <pre className="text-sm whitespace-pre-wrap font-sans leading-relaxed">
                          {chunk}
                        </pre>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Embedding Model
                </label>
                <select
                  value={embedData.model}
                  onChange={(e) => setEmbedData(prev => ({ ...prev, model: e.target.value }))}
                  className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="mistral-embed">mistral-embed</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Provider
                </label>
                <select
                  value={embedData.provider}
                  onChange={(e) => setEmbedData(prev => ({ ...prev, provider: e.target.value }))}
                  className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="mistral">mistral</option>
                </select>
              </div>
            </div>
            <div className="flex space-x-4">
              <button
                onClick={embedDocument}
                disabled={!embedData.input.trim() || isLoading}
                className="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Embedding...' : 
                  embedData.chunkingStrategy !== 'none' ? 
                    `Embed ${chunkText(embedData.input, embedData.chunkingStrategy, embedData.chunkSize, embedData.chunkOverlap).length} Chunks` : 
                    'Embed Content'
                }
              </button>
              <button
                onClick={cancelEmbed}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Documents List */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Documents</h2>
            <button
              onClick={loadDocuments}
              disabled={isLoading}
              className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800 disabled:opacity-50"
            >
              {isLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>
          
          {documents.length === 0 ? (
            <div className="text-gray-500 text-center py-8">
              {isLoading ? 'Loading documents...' : 'No documents found. Create your first document!'}
            </div>
          ) : (
            <div className="space-y-3">
              {documents.map((document) => (
                <div
                  key={document.id}
                  className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedDocument?.id === document.id
                      ? 'bg-blue-50 border-blue-200'
                      : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
                  }`}
                  onClick={() => loadDocumentDetails(document.id)}
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className="font-medium text-gray-900">{document.name}</h3>
                      {document.description && (
                        <p className="text-sm text-gray-600 mt-1">{document.description}</p>
                      )}
                      <p className="text-xs text-gray-500 mt-2">
                        Created: {new Date(document['created-at']).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex space-x-1 ml-2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedDocument(document);
                          setShowEmbedForm(true);
                        }}
                        className="text-green-500 hover:text-green-700"
                        title="Embed content"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          deleteDocument(document.id);
                        }}
                        className="text-red-500 hover:text-red-700"
                        title="Delete document"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Document Details and Chunks */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <h2 className="text-xl font-semibold mb-4">Document Details</h2>
          
          {selectedDocument ? (
            <div className="space-y-6">
              <div>
                <h3 className="font-medium text-gray-900">{selectedDocument.name}</h3>
                {selectedDocument.description && (
                  <p className="text-sm text-gray-600 mt-1">{selectedDocument.description}</p>
                )}
                <p className="text-xs text-gray-500 mt-2">
                  ID: {selectedDocument.id}
                </p>
                <p className="text-xs text-gray-500">
                  Created: {new Date(selectedDocument['created-at']).toLocaleString()}
                </p>
              </div>

              <div className="flex space-x-2">
                <button
                  onClick={() => setShowEmbedForm(true)}
                  className="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600"
                >
                  Add Content
                </button>
                <button
                  onClick={() => deleteDocument(selectedDocument.id)}
                  className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600"
                >
                  Delete Document
                </button>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-3">Document Chunks ({documentChunks.length})</h4>
                {documentChunks.length === 0 ? (
                  <div className="text-gray-500 text-center py-4">
                    No content chunks found. Add content to this document to enable RAG search.
                  </div>
                ) : (
                  <div className="space-y-3 max-h-96 overflow-y-auto">
                    {documentChunks.map((chunk) => (
                      <div key={chunk.id} className="bg-gray-50 p-3 rounded-md">
                        <div className="flex justify-between items-start mb-2">
                          <p className="text-xs text-gray-500">
                            Chunk ID: {chunk.id}
                          </p>
                          <button
                            onClick={() => deleteDocumentChunk(chunk.id)}
                            className="text-red-500 hover:text-red-700"
                            title="Delete chunk"
                          >
                            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                        <p className="text-sm text-gray-700 whitespace-pre-wrap">
                          {chunk.fragment}
                        </p>
                        <p className="text-xs text-gray-500 mt-2">
                          Created: {new Date(chunk['created-at']).toLocaleString()}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div className="text-gray-500 text-center py-8">
              Select a document from the list to view details and manage content
            </div>
          )}
        </div>
      </div>
    </div>
  );
}