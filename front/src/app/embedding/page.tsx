'use client';

import { useState } from 'react';
import { 
  ClientRagSearchQuery,
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

export default function EmbeddingPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<ApiError | null>(null);
  const [searchResults, setSearchResults] = useState<ClientDocumentChunk[]>([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [apiBaseUrl, setApiBaseUrl] = useState(process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080');

  // Search form state
  const [searchQuery, setSearchQuery] = useState({
    input: '',
    limit: 10,
    model: 'mistral-embed',
    provider: 'mistral',
  });

  const performSearch = async () => {
    if (!searchQuery.input.trim()) return;

    setIsLoading(true);
    setApiError(null);
    setHasSearched(false);

    try {
      const ragQuery: ClientRagSearchQuery = {
        input: searchQuery.input.trim(),
        limit: searchQuery.limit,
        model: searchQuery.model,
        provider: searchQuery.provider,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/document-chunk`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(ragQuery),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListDocumentChunksOutput = await response.json();
      setSearchResults(data.chunks || []);
      setHasSearched(true);
    } catch (error) {
      console.error('Error performing embedding search:', error);
      const apiError: ApiError = {
        message: 'Failed to perform embedding search',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
      setSearchResults([]);
      setHasSearched(true);
    } finally {
      setIsLoading(false);
    }
  };

  const dismissError = () => {
    setApiError(null);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    performSearch();
  };

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Embedding Search</h1>
          <p className="text-gray-600 mt-2">Search through embedded document chunks using RAG (Retrieval-Augmented Generation)</p>
        </div>
        <div className="flex space-x-4">
          <input
            type="url"
            value={apiBaseUrl}
            onChange={(e) => setApiBaseUrl(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="API Base URL"
          />
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
              <h3 className="text-sm font-medium text-red-800">Search Error</h3>
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

      {/* Search Form */}
      <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">Search Parameters</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Search Query *
            </label>
            <textarea
              value={searchQuery.input}
              onChange={(e) => setSearchQuery(prev => ({ ...prev, input: e.target.value }))}
              className="w-full p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              rows={4}
              placeholder="Enter your search query to find relevant document chunks..."
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              This query will be embedded and compared against all document chunks to find the most relevant matches.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Max Results
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={searchQuery.limit}
                onChange={(e) => setSearchQuery(prev => ({ ...prev, limit: parseInt(e.target.value) || 10 }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Embedding Model
              </label>
              <select
                value={searchQuery.model}
                onChange={(e) => setSearchQuery(prev => ({ ...prev, model: e.target.value }))}
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
                value={searchQuery.provider}
                onChange={(e) => setSearchQuery(prev => ({ ...prev, provider: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="mistral">mistral</option>
              </select>
            </div>
          </div>

          <div className="flex space-x-4">
            <button
              type="submit"
              disabled={!searchQuery.input.trim() || isLoading}
              className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
            >
              {isLoading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-3 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Searching...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                  Search Embeddings
                </>
              )}
            </button>
            
            {searchResults.length > 0 && (
              <button
                type="button"
                onClick={() => {
                  setSearchResults([]);
                  setHasSearched(false);
                  setSearchQuery(prev => ({ ...prev, input: '' }));
                }}
                className="px-4 py-2 text-gray-600 hover:text-gray-800 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Clear Results
              </button>
            )}
          </div>
        </form>
      </div>

      {/* Search Results */}
      {hasSearched && (
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Search Results</h2>
            <div className="text-sm text-gray-500">
              {searchResults.length} chunk{searchResults.length !== 1 ? 's' : ''} found
            </div>
          </div>

          {searchResults.length === 0 ? (
            <div className="text-center py-12">
              <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 className="text-lg font-medium text-gray-900 mb-2">No matching chunks found</h3>
              <p className="text-gray-500">
                Try adjusting your search query or embedding more documents with relevant content.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {searchResults.map((chunk, index) => (
                <div
                  key={chunk.id}
                  className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex justify-between items-start mb-3">
                    <div className="flex items-center space-x-3">
                      <div className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded-full font-medium">
                        #{index + 1}
                      </div>
                      <div className="text-xs text-gray-500">
                        Document ID: {chunk['document-id']}
                      </div>
                      <div className="text-xs text-gray-500">
                        Chunk ID: {chunk.id}
                      </div>
                    </div>
                    <div className="text-xs text-gray-500">
                      {new Date(chunk['created-at']).toLocaleString()}
                    </div>
                  </div>

                  <div className="text-gray-900">
                    <div className="bg-gray-50 p-3 rounded-md border-l-4 border-blue-400">
                      <p className="text-sm leading-relaxed whitespace-pre-wrap">
                        {chunk.fragment}
                      </p>
                    </div>
                  </div>

                  <div className="mt-3 flex justify-between items-center text-xs text-gray-500">
                    <div>
                      Fragment length: {chunk.fragment.length} characters
                    </div>
                    <div>
                      Created: {new Date(chunk['created-at']).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {searchResults.length > 0 && (
            <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
              <div className="flex items-start">
                <svg className="w-5 h-5 text-blue-500 mt-0.5 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div className="text-sm text-blue-800">
                  <p className="font-medium mb-1">About these results:</p>
                  <p>
                    These document chunks were ranked by semantic similarity to your query using the {searchQuery.model} embedding model. 
                    The most relevant chunks appear first and can be used to provide context for AI conversations.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}