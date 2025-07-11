'use client';

import { useState, useEffect, useCallback } from 'react';
import { 
  ClientContextMetadata, 
  ClientListContextOutput, 
  ClientCreateContextInput, 
  ClientResponse,
  ClientContext,
  ClientNewMessage,
  ClientUpdateContextMessageInput,
  ClientAddMessagesToContextInput
} from '@/types/api';

interface ApiError {
  message: string;
  status?: number;
  details?: string;
  timestamp: string;
}

export default function ContextPage() {
  const [contexts, setContexts] = useState<ClientContextMetadata[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<ApiError | null>(null);
  const [selectedContext, setSelectedContext] = useState<ClientContext | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [apiBaseUrl, setApiBaseUrl] = useState(process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080');

  // Create form state
  const [createForm, setCreateForm] = useState({
    name: '',
    description: '',
  });

  // Message management state
  const [showAddMessage, setShowAddMessage] = useState(false);
  const [editingMessage, setEditingMessage] = useState<string | null>(null);
  const [messageForm, setMessageForm] = useState({
    role: 'user',
    content: '',
  });

  const loadContexts = useCallback(async () => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/context`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListContextOutput = await response.json();
      setContexts(data.contexts || []);
    } catch (error) {
      console.error('Error loading contexts:', error);
      const apiError: ApiError = {
        message: 'Failed to load contexts',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl]);

  const loadContextDetails = async (contextId: string) => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/context/${contextId}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientContext = await response.json();
      setSelectedContext(data);
    } catch (error) {
      console.error('Error loading context details:', error);
      const apiError: ApiError = {
        message: 'Failed to load context details',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const createContext = async () => {
    if (!createForm.name.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const contextInput: ClientCreateContextInput = {
        name: createForm.name,
        description: createForm.description || undefined,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/context`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(contextInput),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to create context');
      }

      // Reset form and reload contexts
      setCreateForm({ name: '', description: '' });
      setShowCreateForm(false);
      await loadContexts();
    } catch (error) {
      console.error('Error creating context:', error);
      const apiError: ApiError = {
        message: 'Failed to create context',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteContext = async (contextId: string) => {
    if (!confirm('Are you sure you want to delete this context? This action cannot be undone.')) {
      return;
    }

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/context/${contextId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete context');
      }

      // If the deleted context was selected, clear selection
      if (selectedContext?.id === contextId) {
        setSelectedContext(null);
      }

      // Reload contexts
      await loadContexts();
    } catch (error) {
      console.error('Error deleting context:', error);
      const apiError: ApiError = {
        message: 'Failed to delete context',
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

  // Message management functions
  const addMessage = async () => {
    if (!selectedContext || !messageForm.content.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const newMessage: ClientNewMessage = {
        role: messageForm.role,
        content: messageForm.content,
      };

      const payload: ClientAddMessagesToContextInput = {
        messages: [newMessage],
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/context/${selectedContext.id}/message`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to add message');
      }

      // Reset form and reload context details
      setMessageForm({ role: 'user', content: '' });
      setShowAddMessage(false);
      await loadContextDetails(selectedContext.id);
    } catch (error) {
      console.error('Error adding message:', error);
      const apiError: ApiError = {
        message: 'Failed to add message',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const updateMessage = async (messageId: string) => {
    if (!messageForm.content.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const updateData: ClientUpdateContextMessageInput = {
        role: messageForm.role,
        content: messageForm.content,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/message/${messageId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updateData),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to update message');
      }

      // Reset form and reload context details
      setMessageForm({ role: 'user', content: '' });
      setEditingMessage(null);
      if (selectedContext) {
        await loadContextDetails(selectedContext.id);
      }
    } catch (error) {
      console.error('Error updating message:', error);
      const apiError: ApiError = {
        message: 'Failed to update message',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteMessage = async (messageId: string) => {
    if (!confirm('Are you sure you want to delete this message?')) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/message/${messageId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete message');
      }

      // Reload context details
      if (selectedContext) {
        await loadContextDetails(selectedContext.id);
      }
    } catch (error) {
      console.error('Error deleting message:', error);
      const apiError: ApiError = {
        message: 'Failed to delete message',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteAllMessages = async () => {
    if (!selectedContext) return;
    if (!confirm('Are you sure you want to delete ALL messages in this context? This action cannot be undone.')) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/context/${selectedContext.id}/message`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete all messages');
      }

      // Reload context details
      await loadContextDetails(selectedContext.id);
    } catch (error) {
      console.error('Error deleting all messages:', error);
      const apiError: ApiError = {
        message: 'Failed to delete all messages',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const startEditMessage = (messageId: string, role: string, content: string) => {
    setEditingMessage(messageId);
    setMessageForm({ role, content });
    setShowAddMessage(false);
  };

  const cancelEdit = () => {
    setEditingMessage(null);
    setMessageForm({ role: 'user', content: '' });
  };

  // Load contexts on component mount
  useEffect(() => {
    loadContexts();
  }, [loadContexts]);

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Context Management</h1>
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
            {showCreateForm ? 'Cancel' : 'Create Context'}
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

      {/* Create Context Form */}
      {showCreateForm && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Create New Context</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Context Name *
              </label>
              <input
                type="text"
                value={createForm.name}
                onChange={(e) => setCreateForm(prev => ({ ...prev, name: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter context name"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                value={createForm.description}
                onChange={(e) => setCreateForm(prev => ({ ...prev, description: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                rows={3}
                placeholder="Enter context description (optional)"
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={createContext}
                disabled={!createForm.name.trim() || isLoading}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Creating...' : 'Create Context'}
              </button>
              <button
                onClick={() => setShowCreateForm(false)}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Context List */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Contexts</h2>
            <button
              onClick={loadContexts}
              disabled={isLoading}
              className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800 disabled:opacity-50"
            >
              {isLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>
          
          {contexts.length === 0 ? (
            <div className="text-gray-500 text-center py-8">
              {isLoading ? 'Loading contexts...' : 'No contexts found. Create your first context!'}
            </div>
          ) : (
            <div className="space-y-3">
              {contexts.map((context) => (
                <div
                  key={context.id}
                  className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedContext?.id === context.id
                      ? 'bg-blue-50 border-blue-200'
                      : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
                  }`}
                  onClick={() => loadContextDetails(context.id)}
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className="font-medium text-gray-900">{context.name}</h3>
                      {context.description && (
                        <p className="text-sm text-gray-600 mt-1">{context.description}</p>
                      )}
                      <p className="text-xs text-gray-500 mt-2">
                        Created: {new Date(context['created-at']).toLocaleString()}
                      </p>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        deleteContext(context.id);
                      }}
                      className="ml-2 text-red-500 hover:text-red-700"
                      title="Delete context"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Context Details */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <h2 className="text-xl font-semibold mb-4">Context Details</h2>
          
          {selectedContext ? (
            <div className="space-y-4">
              <div>
                <h3 className="font-medium text-gray-900">{selectedContext.name}</h3>
                {selectedContext.description && (
                  <p className="text-sm text-gray-600 mt-1">{selectedContext.description}</p>
                )}
                <p className="text-xs text-gray-500 mt-2">
                  ID: {selectedContext.id}
                </p>
                <p className="text-xs text-gray-500">
                  Created: {new Date(selectedContext['created-at']).toLocaleString()}
                </p>
              </div>

              {/* Messages */}
              <div>
                <div className="flex justify-between items-center mb-2">
                  <h4 className="font-medium text-gray-900">Messages</h4>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => setShowAddMessage(!showAddMessage)}
                      className="px-3 py-1 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600"
                    >
                      {showAddMessage ? 'Cancel' : 'Add Message'}
                    </button>
                    {selectedContext.messages && selectedContext.messages.length > 0 && (
                      <button
                        onClick={deleteAllMessages}
                        className="px-3 py-1 text-sm bg-red-500 text-white rounded-md hover:bg-red-600"
                      >
                        Clear All
                      </button>
                    )}
                  </div>
                </div>

                {/* Add/Edit Message Form */}
                {(showAddMessage || editingMessage) && (
                  <div className="mb-4 p-3 bg-gray-50 rounded-md">
                    <div className="space-y-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Role
                        </label>
                        <select
                          value={messageForm.role}
                          onChange={(e) => setMessageForm(prev => ({ ...prev, role: e.target.value }))}
                          className="w-full p-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        >
                          <option value="user">User</option>
                          <option value="assistant">Assistant</option>
                          <option value="system">System</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Content
                        </label>
                        <textarea
                          value={messageForm.content}
                          onChange={(e) => setMessageForm(prev => ({ ...prev, content: e.target.value }))}
                          className="w-full p-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          rows={3}
                          placeholder="Enter message content"
                        />
                      </div>
                      <div className="flex space-x-2">
                        {editingMessage ? (
                          <>
                            <button
                              onClick={() => updateMessage(editingMessage)}
                              disabled={!messageForm.content.trim() || isLoading}
                              className="px-3 py-1 text-sm bg-green-500 text-white rounded-md hover:bg-green-600 disabled:opacity-50"
                            >
                              {isLoading ? 'Updating...' : 'Update'}
                            </button>
                            <button
                              onClick={cancelEdit}
                              className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
                            >
                              Cancel
                            </button>
                          </>
                        ) : (
                          <>
                            <button
                              onClick={addMessage}
                              disabled={!messageForm.content.trim() || isLoading}
                              className="px-3 py-1 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50"
                            >
                              {isLoading ? 'Adding...' : 'Add Message'}
                            </button>
                            <button
                              onClick={() => {
                                setShowAddMessage(false);
                                setMessageForm({ role: 'user', content: '' });
                              }}
                              className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
                            >
                              Cancel
                            </button>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {/* Messages List */}
                {selectedContext.messages && selectedContext.messages.length > 0 ? (
                  <div className="space-y-2 max-h-60 overflow-y-auto">
                    {selectedContext.messages.map((message) => (
                      <div
                        key={message.id}
                        className={`p-3 rounded-md text-sm border ${
                          message.role === 'user'
                            ? 'bg-blue-100 text-blue-900 border-blue-200'
                            : message.role === 'assistant'
                            ? 'bg-green-100 text-green-900 border-green-200'
                            : 'bg-gray-100 text-gray-900 border-gray-200'
                        }`}
                      >
                        <div className="flex justify-between items-start">
                          <div className="flex-1">
                            <div className="font-medium capitalize mb-1">{message.role}</div>
                            <div className="whitespace-pre-wrap">{message.content}</div>
                            <div className="text-xs text-gray-500 mt-1">
                              {new Date(message['created-at']).toLocaleString()}
                            </div>
                          </div>
                          <div className="flex space-x-1 ml-2">
                            <button
                              onClick={() => startEditMessage(message.id, message.role, message.content)}
                              className="p-1 text-gray-500 hover:text-blue-600"
                              title="Edit message"
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                              </svg>
                            </button>
                            <button
                              onClick={() => deleteMessage(message.id)}
                              className="p-1 text-gray-500 hover:text-red-600"
                              title="Delete message"
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
                ) : (
                  <p className="text-gray-500 text-sm">No messages in this context.</p>
                )}
              </div>

              {/* Sources */}
              {selectedContext.sources && selectedContext.sources.contexts && (
                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Source Contexts</h4>
                  {selectedContext.sources.contexts.length > 0 ? (
                    <div className="space-y-1">
                      {selectedContext.sources.contexts.map((sourceId) => (
                        <div key={sourceId} className="text-sm text-gray-600 bg-gray-50 p-2 rounded">
                          {sourceId}
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-gray-500 text-sm">No source contexts.</p>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="text-gray-500 text-center py-8">
              Select a context from the list to view details
            </div>
          )}
        </div>
      </div>
    </div>
  );
}