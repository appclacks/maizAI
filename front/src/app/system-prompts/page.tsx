'use client';

import { useState, useEffect, useCallback } from 'react';
import { 
  ClientSystemPrompt, 
  ClientListSystemPromptsOutput, 
  ClientCreateSystemPromptInput, 
  ClientUpdateSystemPromptInput,
  ClientResponse 
} from '@/types/api';

interface ApiError {
  message: string;
  status?: number;
  details?: string;
  timestamp: string;
}

export default function SystemPromptsPage() {
  const [systemPrompts, setSystemPrompts] = useState<ClientSystemPrompt[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [apiError, setApiError] = useState<ApiError | null>(null);
  const [selectedPrompt, setSelectedPrompt] = useState<ClientSystemPrompt | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingPrompt, setEditingPrompt] = useState<string | null>(null);
  const [apiBaseUrl, setApiBaseUrl] = useState(process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080');

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    content: '',
  });

  const loadSystemPrompts = useCallback(async () => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/system-prompt`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListSystemPromptsOutput = await response.json();
      setSystemPrompts(data.system_prompts || []);
    } catch (error) {
      console.error('Error loading system prompts:', error);
      const apiError: ApiError = {
        message: 'Failed to load system prompts',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl]);

  const loadPromptDetails = async (promptId: string) => {
    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/system-prompt/${promptId}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientSystemPrompt = await response.json();
      setSelectedPrompt(data);
    } catch (error) {
      console.error('Error loading system prompt details:', error);
      const apiError: ApiError = {
        message: 'Failed to load system prompt details',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const createSystemPrompt = async () => {
    if (!formData.name.trim() || !formData.content.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const promptInput: ClientCreateSystemPromptInput = {
        name: formData.name,
        description: formData.description || undefined,
        content: formData.content,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/system-prompt`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(promptInput),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to create system prompt');
      }

      // Reset form and reload prompts
      setFormData({ name: '', description: '', content: '' });
      setShowCreateForm(false);
      await loadSystemPrompts();
    } catch (error) {
      console.error('Error creating system prompt:', error);
      const apiError: ApiError = {
        message: 'Failed to create system prompt',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const updateSystemPrompt = async (promptId: string) => {
    if (!formData.content.trim()) return;

    setIsLoading(true);
    setApiError(null);

    try {
      const updateData: ClientUpdateSystemPromptInput = {
        content: formData.content,
      };

      const response = await fetch(`${apiBaseUrl}/api/v1/system-prompt/${promptId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updateData),
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to update system prompt');
      }

      // Reset form and reload prompts
      setFormData({ name: '', description: '', content: '' });
      setEditingPrompt(null);
      await loadSystemPrompts();
      if (selectedPrompt) {
        await loadPromptDetails(selectedPrompt.id);
      }
    } catch (error) {
      console.error('Error updating system prompt:', error);
      const apiError: ApiError = {
        message: 'Failed to update system prompt',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };
      setApiError(apiError);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteSystemPrompt = async (promptId: string) => {
    if (!confirm('Are you sure you want to delete this system prompt? This action cannot be undone.')) {
      return;
    }

    setIsLoading(true);
    setApiError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/system-prompt/${promptId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData: ClientResponse = await response.json();
        throw new Error(errorData.messages?.join(', ') || 'Failed to delete system prompt');
      }

      // If the deleted prompt was selected, clear selection
      if (selectedPrompt?.id === promptId) {
        setSelectedPrompt(null);
      }

      // Reload prompts
      await loadSystemPrompts();
    } catch (error) {
      console.error('Error deleting system prompt:', error);
      const apiError: ApiError = {
        message: 'Failed to delete system prompt',
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

  const startEdit = (prompt: ClientSystemPrompt) => {
    setEditingPrompt(prompt.id);
    setFormData({
      name: prompt.name,
      description: prompt.description || '',
      content: prompt.content,
    });
    setShowCreateForm(false);
  };

  const cancelEdit = () => {
    setEditingPrompt(null);
    setFormData({ name: '', description: '', content: '' });
  };

  const cancelCreate = () => {
    setShowCreateForm(false);
    setFormData({ name: '', description: '', content: '' });
  };

  // Load system prompts on component mount
  useEffect(() => {
    loadSystemPrompts();
  }, [loadSystemPrompts]);

  return (
    <div className="max-w-6xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">System Prompt Management</h1>
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
            {showCreateForm ? 'Cancel' : 'Create System Prompt'}
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
          <h2 className="text-xl font-semibold mb-4">Create New System Prompt</h2>
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
                placeholder="Enter system prompt name"
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
                rows={2}
                placeholder="Enter system prompt description (optional)"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Content *
              </label>
              <textarea
                value={formData.content}
                onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                rows={8}
                placeholder="Enter system prompt content"
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={createSystemPrompt}
                disabled={!formData.name.trim() || !formData.content.trim() || isLoading}
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Creating...' : 'Create System Prompt'}
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

      {/* Edit Form */}
      {editingPrompt && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Edit System Prompt</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Name (read-only)
              </label>
              <input
                type="text"
                value={formData.name}
                className="w-full p-2 border border-gray-300 rounded-md bg-gray-100 cursor-not-allowed"
                disabled
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Description (read-only)
              </label>
              <textarea
                value={formData.description}
                className="w-full p-2 border border-gray-300 rounded-md bg-gray-100 cursor-not-allowed"
                rows={2}
                disabled
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Content *
              </label>
              <textarea
                value={formData.content}
                onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                rows={8}
                placeholder="Enter system prompt content"
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={() => updateSystemPrompt(editingPrompt)}
                disabled={!formData.content.trim() || isLoading}
                className="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Updating...' : 'Update System Prompt'}
              </button>
              <button
                onClick={cancelEdit}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* System Prompts List */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">System Prompts</h2>
            <button
              onClick={loadSystemPrompts}
              disabled={isLoading}
              className="px-3 py-1 text-sm text-gray-600 hover:text-gray-800 disabled:opacity-50"
            >
              {isLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>
          
          {systemPrompts.length === 0 ? (
            <div className="text-gray-500 text-center py-8">
              {isLoading ? 'Loading system prompts...' : 'No system prompts found. Create your first prompt!'}
            </div>
          ) : (
            <div className="space-y-3">
              {systemPrompts.map((prompt) => (
                <div
                  key={prompt.id}
                  className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedPrompt?.id === prompt.id
                      ? 'bg-blue-50 border-blue-200'
                      : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
                  }`}
                  onClick={() => loadPromptDetails(prompt.id)}
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className="font-medium text-gray-900">{prompt.name}</h3>
                      {prompt.description && (
                        <p className="text-sm text-gray-600 mt-1">{prompt.description}</p>
                      )}
                      <p className="text-xs text-gray-500 mt-2">
                        Created: {new Date(prompt['created-at']).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex space-x-1 ml-2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          startEdit(prompt);
                        }}
                        className="text-blue-500 hover:text-blue-700"
                        title="Edit system prompt"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          deleteSystemPrompt(prompt.id);
                        }}
                        className="text-red-500 hover:text-red-700"
                        title="Delete system prompt"
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

        {/* System Prompt Details */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <h2 className="text-xl font-semibold mb-4">System Prompt Details</h2>
          
          {selectedPrompt ? (
            <div className="space-y-4">
              <div>
                <h3 className="font-medium text-gray-900">{selectedPrompt.name}</h3>
                {selectedPrompt.description && (
                  <p className="text-sm text-gray-600 mt-1">{selectedPrompt.description}</p>
                )}
                <p className="text-xs text-gray-500 mt-2">
                  ID: {selectedPrompt.id}
                </p>
                <p className="text-xs text-gray-500">
                  Created: {new Date(selectedPrompt['created-at']).toLocaleString()}
                </p>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">Content</h4>
                <div className="bg-gray-50 p-4 rounded-md max-h-96 overflow-y-auto">
                  <pre className="text-sm text-gray-700 whitespace-pre-wrap font-mono">
                    {selectedPrompt.content}
                  </pre>
                </div>
              </div>

              <div className="flex space-x-2">
                <button
                  onClick={() => startEdit(selectedPrompt)}
                  className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
                >
                  Edit Content
                </button>
                <button
                  onClick={() => deleteSystemPrompt(selectedPrompt.id)}
                  className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600"
                >
                  Delete
                </button>
              </div>
            </div>
          ) : (
            <div className="text-gray-500 text-center py-8">
              Select a system prompt from the list to view details
            </div>
          )}
        </div>
      </div>
    </div>
  );
}