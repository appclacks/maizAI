'use client';

import { useState, useEffect, useCallback } from 'react';
import { ClientCreateConversationInput, ClientConversationAnswer, ClientNewMessage, ClientContextMetadata, ClientListContextOutput, ClientSystemPrompt, ClientListSystemPromptsOutput, ClientContext } from '@/types/api';

interface ApiError {
  message: string;
  status?: number;
  details?: string;
  timestamp: string;
}

export default function ConversationPage() {
  const [messages, setMessages] = useState<ClientNewMessage[]>([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [contextId, setContextId] = useState<string | null>(null);
  const [response, setResponse] = useState<ClientConversationAnswer | null>(null);
  const [apiError, setApiError] = useState<ApiError | null>(null);
  
  // Default configuration
  const defaultConfig = {
    model: 'claude-3-7-sonnet-latest',
    provider: 'anthropic',
    maxTokens: 1000,
    temperature: 0.7,
    contextName: 'New Conversation',
    systemPrompt: '',
    apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  };

  // RAG configuration
  const [ragEnabled, setRagEnabled] = useState(false);
  const [showRagConfig, setShowRagConfig] = useState(false);
  const [ragConfig, setRagConfig] = useState({
    input: '',
    limit: 5,
    model: 'mistral-embed',
    provider: 'mistral',
  });

  // Configuration options
  const [config, setConfig] = useState(defaultConfig);

  // Context selection
  const [availableContexts, setAvailableContexts] = useState<ClientContextMetadata[]>([]);
  const [selectedContextId, setSelectedContextId] = useState<string>('');
  const [useNewContext, setUseNewContext] = useState(true);
  const [contextDescription, setContextDescription] = useState('');
  const [selectedSourceContexts, setSelectedSourceContexts] = useState<string[]>([]);

  // System prompt selection
  const [availableSystemPrompts, setAvailableSystemPrompts] = useState<ClientSystemPrompt[]>([]);
  const [selectedSystemPromptId, setSelectedSystemPromptId] = useState<string>('');
  const [useSystemPromptId, setUseSystemPromptId] = useState(false);

  const sendMessage = async () => {
    if (!currentMessage.trim()) return;

    const userMessage: ClientNewMessage = {
      role: 'user',
      content: currentMessage,
    };

    const newMessages = [...messages, userMessage];
    setMessages(newMessages);
    setCurrentMessage('');
    setIsLoading(true);
    setApiError(null); // Clear any previous errors

    try {
      const conversationInput: ClientCreateConversationInput = {
        messages: [userMessage],
        'context-id': contextId || (!useNewContext && selectedContextId ? selectedContextId : undefined),
        'new-context': (!contextId && useNewContext) ? {
          name: config.contextName,
          description: contextDescription.trim() || undefined,
          sources: selectedSourceContexts.length > 0 ? {
            contexts: selectedSourceContexts,
          } : undefined,
        } : undefined,
        'system-prompt-id': useSystemPromptId && selectedSystemPromptId ? selectedSystemPromptId : undefined,
        'query-options': {
          model: config.model,
          'max-tokens': config.maxTokens,
          provider: config.provider,
          temperature: config.temperature,
          system: !useSystemPromptId && config.systemPrompt ? config.systemPrompt : undefined,
          rag: ragEnabled && ragConfig.input.trim() ? {
            input: ragConfig.input,
            limit: ragConfig.limit,
            model: ragConfig.model,
            provider: ragConfig.provider,
          } : undefined,
        },
      };

      const apiUrl = `${config.apiBaseUrl}/api/v1/conversation`;
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(conversationInput),
      });

      if (!response.ok) {
        const errorMessage = 'Failed to send message';
        let errorDetails = '';
        
        try {
          const errorData = await response.json();
          if (errorData.messages && Array.isArray(errorData.messages)) {
            errorDetails = errorData.messages.join(', ');
          } else if (errorData.message) {
            errorDetails = errorData.message;
          }
        } catch {
          // If we can't parse the error response, use the status text
          errorDetails = response.statusText || 'Unknown error';
        }

        const apiError: ApiError = {
          message: errorMessage,
          status: response.status,
          details: errorDetails,
          timestamp: new Date().toISOString(),
        };

        setApiError(apiError);
        
        // Add error message to chat
        const errorChatMessage: ClientNewMessage = {
          role: 'assistant',
          content: `API Error (${response.status}): ${errorDetails || 'Unknown error occurred'}`,
        };
        setMessages(prev => [...prev, errorChatMessage]);
        return;
      }

      const data: ClientConversationAnswer = await response.json();
      setResponse(data);
      
      // Set context ID for future messages
      if (data.context) {
        setContextId(data.context);
      }

      // Add assistant response to messages
      if (data.result && data.result.length > 0) {
        const assistantMessage: ClientNewMessage = {
          role: 'assistant',
          content: data.result.map(r => r.text).join(' '),
        };
        setMessages(prev => [...prev, assistantMessage]);
      }
    } catch (error) {
      console.error('Error sending message:', error);
      
      // Create detailed error information
      const apiError: ApiError = {
        message: 'Network or connection error',
        details: error instanceof Error ? error.message : 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      };

      setApiError(apiError);
      
      // Add error message to chat
      const connectionErrorMessage: ClientNewMessage = {
        role: 'assistant',
        content: `Connection Error: ${error instanceof Error ? error.message : 'Unable to connect to the API'}`,
      };
      setMessages(prev => [...prev, connectionErrorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const clearConversation = () => {
    setMessages([]);
    setContextId(null);
    setResponse(null);
    setApiError(null);
  };

  const startNewConversation = () => {
    setMessages([]);
    setContextId(null);
    setResponse(null);
    setApiError(null);
    setConfig(defaultConfig);
    setRagEnabled(false);
    setShowRagConfig(false);
    setRagConfig({
      input: '',
      limit: 5,
      model: 'mistral-embed',
      provider: 'mistral',
    });
    setUseNewContext(true);
    setSelectedContextId('');
    setContextDescription('');
    setSelectedSourceContexts([]);
    setUseSystemPromptId(false);
    setSelectedSystemPromptId('');
  };

  const dismissError = () => {
    setApiError(null);
  };

  const copyErrorDetails = () => {
    if (apiError) {
      const errorReport = `
API Error Report
================
Timestamp: ${apiError.timestamp}
Status: ${apiError.status || 'N/A'}
Message: ${apiError.message}
Details: ${apiError.details || 'None'}

Configuration:
- API Base URL: ${config.apiBaseUrl}
- Model: ${config.model}
- Provider: ${config.provider}
- Max Tokens: ${config.maxTokens}
- Temperature: ${config.temperature}
- Context Name: ${config.contextName}
- System Prompt: ${config.systemPrompt || 'None'}

Context ID: ${contextId || 'None'}
      `.trim();
      
      navigator.clipboard.writeText(errorReport).then(() => {
        alert('Error details copied to clipboard!');
      }).catch(() => {
        console.error('Failed to copy error details');
        alert('Failed to copy error details to clipboard');
      });
    }
  };

  const loadAvailableContexts = useCallback(async () => {
    try {
      const response = await fetch(`${config.apiBaseUrl}/api/v1/context`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListContextOutput = await response.json();
      setAvailableContexts(data.contexts || []);
    } catch (error) {
      console.error('Error loading available contexts:', error);
      setAvailableContexts([]);
    }
  }, [config.apiBaseUrl]);

  const loadContextMessages = useCallback(async (contextId: string) => {
    try {
      const response = await fetch(`${config.apiBaseUrl}/api/v1/context/${contextId}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientContext = await response.json();
      if (data.messages) {
        const convertedMessages: ClientNewMessage[] = data.messages.map(msg => ({
          role: msg.role,
          content: msg.content,
        }));
        setMessages(convertedMessages);
      }
    } catch (error) {
      console.error('Error loading context messages:', error);
    }
  }, [config.apiBaseUrl]);

  const loadAvailableSystemPrompts = useCallback(async () => {
    try {
      const response = await fetch(`${config.apiBaseUrl}/api/v1/system-prompt`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data: ClientListSystemPromptsOutput = await response.json();
      setAvailableSystemPrompts(data.system_prompts || []);
    } catch (error) {
      console.error('Error loading available system prompts:', error);
      setAvailableSystemPrompts([]);
    }
  }, [config.apiBaseUrl]);

  // Load available contexts and system prompts on component mount
  useEffect(() => {
    loadAvailableContexts();
    loadAvailableSystemPrompts();
  }, [loadAvailableContexts, loadAvailableSystemPrompts]);

  return (
    <div className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-900 mb-6">AI Conversation</h1>
        
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
                <h3 className="text-sm font-medium text-red-800">
                  API Error Occurred
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p><strong>Message:</strong> {apiError.message}</p>
                  {apiError.status && (
                    <p><strong>Status:</strong> {apiError.status}</p>
                  )}
                  {apiError.details && (
                    <p><strong>Details:</strong> {apiError.details}</p>
                  )}
                  <p><strong>Time:</strong> {new Date(apiError.timestamp).toLocaleString()}</p>
                </div>
                <div className="mt-4 flex space-x-2">
                  <button
                    type="button"
                    onClick={copyErrorDetails}
                    className="bg-red-100 text-red-800 px-3 py-1 rounded-md text-sm font-medium hover:bg-red-200"
                  >
                    Copy Error Report
                  </button>
                  <button
                    type="button"
                    onClick={dismissError}
                    className="bg-red-100 text-red-800 px-3 py-1 rounded-md text-sm font-medium hover:bg-red-200"
                  >
                    Dismiss
                  </button>
                </div>
                <div className="mt-2 text-xs text-red-600">
                  <p><strong>Troubleshooting:</strong></p>
                  <ul className="list-disc list-inside mt-1">
                    <li>Check if the API server is running at: {config.apiBaseUrl}</li>
                    <li>Verify the API Base URL is correct (should not end with /)</li>
                    <li>Ensure all required configuration parameters are valid</li>
                    <li>Check network connectivity and CORS settings</li>
                    <li>Test the API endpoint: {config.apiBaseUrl}/api/v1/conversation</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        )}
        
        {/* Configuration Panel */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Configuration</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="md:col-span-2 lg:col-span-3">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                API Base URL
              </label>
              <input
                type="url"
                value={config.apiBaseUrl}
                onChange={(e) => setConfig(prev => ({ ...prev, apiBaseUrl: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="e.g., http://localhost:8080 or https://api.maizai.com"
              />
              <p className="mt-1 text-xs text-gray-500">
                The base URL for the MaizAI API server (without trailing slash)
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Provider
              </label>
              <select
                value={config.provider}
                onChange={(e) => setConfig(prev => ({ ...prev, provider: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="anthropic">Anthropic</option>
                <option value="mistral">Mistral</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Model
              </label>
              <input
                type="text"
                value={config.model}
                onChange={(e) => setConfig(prev => ({ ...prev, model: e.target.value }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="e.g., gpt-4"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Max Tokens
              </label>
              <input
                type="number"
                value={config.maxTokens}
                onChange={(e) => setConfig(prev => ({ ...prev, maxTokens: parseInt(e.target.value) }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="1000"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Temperature
              </label>
              <input
                type="number"
                step="0.1"
                min="0"
                max="2"
                value={config.temperature}
                onChange={(e) => setConfig(prev => ({ ...prev, temperature: parseFloat(e.target.value) }))}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="0.7"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Context Selection
                {contextId && (
                  <span className="text-xs text-gray-500 ml-1">(locked during conversation)</span>
                )}
              </label>
              <div className="space-y-3">
                <div className="flex items-center space-x-4">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="contextType"
                      checked={useNewContext}
                      onChange={(e) => {
                        const useNew = e.target.checked;
                        setUseNewContext(useNew);
                        if (useNew) {
                          setSelectedContextId('');
                          setMessages([]);
                        }
                      }}
                      className="text-blue-600 focus:ring-blue-500"
                      disabled={!!contextId}
                    />
                    <span className="ml-2 text-sm text-gray-700">Create New Context</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="contextType"
                      checked={!useNewContext}
                      onChange={(e) => {
                        const useExisting = e.target.checked;
                        setUseNewContext(!useExisting);
                        if (!useExisting) {
                          setSelectedContextId('');
                          setMessages([]);
                        }
                      }}
                      className="text-blue-600 focus:ring-blue-500"
                      disabled={!!contextId}
                    />
                    <span className="ml-2 text-sm text-gray-700">Use Existing Context</span>
                  </label>
                </div>
                
                {useNewContext ? (
                  <div className="space-y-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        New Context Name
                      </label>
                      <input
                        type="text"
                        value={config.contextName}
                        onChange={(e) => setConfig(prev => ({ ...prev, contextName: e.target.value }))}
                        className={`w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                          contextId ? 'bg-gray-100 cursor-not-allowed' : ''
                        }`}
                        placeholder="New Conversation"
                        disabled={!!contextId}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Context Description
                      </label>
                      <textarea
                        value={contextDescription}
                        onChange={(e) => setContextDescription(e.target.value)}
                        className={`w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                          contextId ? 'bg-gray-100 cursor-not-allowed' : ''
                        }`}
                        rows={2}
                        placeholder="Optional description for the context"
                        disabled={!!contextId}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Source Contexts
                      </label>
                      <div className="space-y-2">
                        {availableContexts.length > 0 ? (
                          <div className="max-h-32 overflow-y-auto border border-gray-300 rounded-md p-2">
                            {availableContexts.map((context) => (
                              <label key={context.id} className="flex items-center space-x-2 py-1">
                                <input
                                  type="checkbox"
                                  checked={selectedSourceContexts.includes(context.id)}
                                  onChange={(e) => {
                                    if (e.target.checked) {
                                      setSelectedSourceContexts(prev => [...prev, context.id]);
                                    } else {
                                      setSelectedSourceContexts(prev => prev.filter(id => id !== context.id));
                                    }
                                  }}
                                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                  disabled={!!contextId}
                                />
                                <span className="text-sm text-gray-700">{context.name}</span>
                              </label>
                            ))}
                          </div>
                        ) : (
                          <p className="text-xs text-gray-500">No contexts available to use as sources</p>
                        )}
                        {selectedSourceContexts.length > 0 && (
                          <div className="flex flex-wrap gap-1">
                            {selectedSourceContexts.map((sourceId) => {
                              const context = availableContexts.find(c => c.id === sourceId);
                              return context ? (
                                <span
                                  key={sourceId}
                                  className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                                >
                                  {context.name}
                                  <button
                                    onClick={() => setSelectedSourceContexts(prev => prev.filter(id => id !== sourceId))}
                                    className="ml-1 text-blue-600 hover:text-blue-800"
                                    disabled={!!contextId}
                                  >
                                    ×
                                  </button>
                                </span>
                              ) : null;
                            })}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Select Existing Context
                    </label>
                    <div className="flex space-x-2">
                      <select
                        value={selectedContextId}
                        onChange={(e) => {
                          const newContextId = e.target.value;
                          setSelectedContextId(newContextId);
                          if (newContextId) {
                            loadContextMessages(newContextId);
                          } else {
                            setMessages([]);
                          }
                        }}
                        className={`flex-1 p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                          contextId ? 'bg-gray-100 cursor-not-allowed' : ''
                        }`}
                        disabled={!!contextId}
                      >
                        <option value="">Select a context...</option>
                        {availableContexts.map((context) => (
                          <option key={context.id} value={context.id}>
                            {context.name}
                          </option>
                        ))}
                      </select>
                      <button
                        onClick={loadAvailableContexts}
                        className="px-3 py-2 text-sm text-gray-600 hover:text-gray-800 border border-gray-300 rounded-md hover:bg-gray-50"
                        disabled={!!contextId}
                        title="Refresh context list"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                        </svg>
                      </button>
                    </div>
                    {availableContexts.length === 0 && (
                      <p className="mt-1 text-xs text-gray-500">
                        No contexts available. Create your first context or refresh the list.
                      </p>
                    )}
                  </div>
                )}
              </div>
              {contextId && (
                <p className="mt-1 text-xs text-gray-500">
                  Context is locked during conversation. Use &quot;New Conversation&quot; to start fresh.
                </p>
              )}
            </div>
            <div className="md:col-span-2 lg:col-span-3">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                System Prompt
              </label>
              <div className="space-y-3">
                <div className="flex items-center space-x-4">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="systemPromptType"
                      checked={!useSystemPromptId}
                      onChange={(e) => setUseSystemPromptId(!e.target.checked)}
                      className="text-blue-600 focus:ring-blue-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">Custom Text</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="systemPromptType"
                      checked={useSystemPromptId}
                      onChange={(e) => setUseSystemPromptId(e.target.checked)}
                      className="text-blue-600 focus:ring-blue-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">Select System Prompt</span>
                  </label>
                </div>
                
                {!useSystemPromptId ? (
                  <textarea
                    value={config.systemPrompt}
                    onChange={(e) => setConfig(prev => ({ ...prev, systemPrompt: e.target.value }))}
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    rows={3}
                    placeholder="Enter system prompt (optional)"
                  />
                ) : (
                  <div>
                    <div className="flex space-x-2">
                      <select
                        value={selectedSystemPromptId}
                        onChange={(e) => setSelectedSystemPromptId(e.target.value)}
                        className="flex-1 p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        <option value="">Select a system prompt...</option>
                        {availableSystemPrompts.map((prompt) => (
                          <option key={prompt.id} value={prompt.id}>
                            {prompt.name}
                          </option>
                        ))}
                      </select>
                      <button
                        onClick={loadAvailableSystemPrompts}
                        className="px-3 py-2 text-sm text-gray-600 hover:text-gray-800 border border-gray-300 rounded-md hover:bg-gray-50"
                        title="Refresh system prompt list"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                        </svg>
                      </button>
                    </div>
                    {availableSystemPrompts.length === 0 && (
                      <p className="mt-1 text-xs text-gray-500">
                        No system prompts available. Create your first system prompt or refresh the list.
                      </p>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>
          
          {/* RAG Configuration */}
          <div className="mt-6 border-t pt-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-3">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={ragEnabled}
                    onChange={(e) => setRagEnabled(e.target.checked)}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="ml-2 text-sm font-medium text-gray-700">Enable RAG (Retrieval-Augmented Generation)</span>
                </label>
              </div>
              {ragEnabled && (
                <button
                  onClick={() => setShowRagConfig(!showRagConfig)}
                  className="text-sm text-blue-600 hover:text-blue-800 flex items-center"
                >
                  {showRagConfig ? 'Hide' : 'Show'} RAG Configuration
                  <svg 
                    className={`ml-1 h-4 w-4 transition-transform ${showRagConfig ? 'rotate-180' : ''}`}
                    fill="none" 
                    stroke="currentColor" 
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </button>
              )}
            </div>
            
            {ragEnabled && showRagConfig && (
              <div className="bg-gray-50 rounded-lg p-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      RAG Query Input
                    </label>
                    <textarea
                      value={ragConfig.input}
                      onChange={(e) => setRagConfig(prev => ({ ...prev, input: e.target.value }))}
                      className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      rows={3}
                      placeholder="Enter search query for document retrieval (optional)"
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Query used to search relevant documents. Leave empty to use the user message.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Results Limit
                    </label>
                    <input
                      type="number"
                      min="1"
                      max="20"
                      value={ragConfig.limit}
                      onChange={(e) => setRagConfig(prev => ({ ...prev, limit: parseInt(e.target.value) }))}
                      className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="5"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Embedding Provider
                    </label>
                    <select
                      value={ragConfig.provider}
                      onChange={(e) => setRagConfig(prev => ({ ...prev, provider: e.target.value }))}
                      className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="mistral">Mistral</option>
                    </select>
                  </div>
                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Embedding Model
                    </label>
                    <input
                      type="text"
                      value={ragConfig.model}
                      onChange={(e) => setRagConfig(prev => ({ ...prev, model: e.target.value }))}
                      className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="mistral-embed"
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Chat Messages */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <div className="flex justify-between items-center mb-4">
            <div>
              <h2 className="text-xl font-semibold">Conversation</h2>
              {contextId && (
                <p className="text-sm text-gray-500 mt-1">
                  Context ID: {contextId}
                </p>
              )}
            </div>
            <div className="flex space-x-2">
              <button
                onClick={clearConversation}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-md"
              >
                Clear
              </button>
              <button
                onClick={startNewConversation}
                className="px-4 py-2 text-sm bg-blue-500 text-white hover:bg-blue-600 rounded-md"
              >
                New Conversation
              </button>
            </div>
          </div>
          
          {messages.length === 0 ? (
            <div className="text-gray-500 text-center py-8">
              No messages yet. Start a conversation below!
            </div>
          ) : (
            <div className="space-y-4 max-h-96 overflow-y-auto">
              {messages.map((message, index) => (
                <div
                  key={index}
                  className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                      message.role === 'user'
                        ? 'bg-blue-500 text-white'
                        : message.content.includes('Error:') || message.content.includes('API Error')
                        ? 'bg-red-100 text-red-900 border border-red-300'
                        : 'bg-gray-100 text-gray-900'
                    }`}
                  >
                    <div className="text-sm font-medium mb-1 capitalize flex items-center">
                      {message.role}
                      {(message.content.includes('Error:') || message.content.includes('API Error')) && (
                        <svg className="h-4 w-4 ml-1 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                        </svg>
                      )}
                    </div>
                    <div className="whitespace-pre-wrap">{message.content}</div>
                  </div>
                </div>
              ))}
              {isLoading && (
                <div className="flex justify-start">
                  <div className="max-w-xs lg:max-w-md px-4 py-2 rounded-lg bg-gray-100 text-gray-900">
                    <div className="text-sm font-medium mb-1">Assistant</div>
                    <div className="animate-pulse">Thinking...</div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Message Input */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex space-x-4">
            <input
              type="text"
              value={currentMessage}
              onChange={(e) => setCurrentMessage(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && !isLoading && sendMessage()}
              placeholder="Type your message..."
              className="flex-1 p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              disabled={isLoading}
            />
            <button
              onClick={sendMessage}
              disabled={isLoading || !currentMessage.trim()}
              className="px-6 py-3 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Sending...' : 'Send'}
            </button>
          </div>
        </div>

        {/* Response Info */}
        {response && (
          <div className="mt-6 bg-gray-100 rounded-lg p-4">
            <h3 className="font-semibold text-gray-900 mb-2">Response Info</h3>
            <div className="text-sm text-gray-600 space-y-1">
              <div>Context ID: {response.context}</div>
              <div>Input Tokens: {response['input-tokens']}</div>
              <div>Output Tokens: {response['output-tokens']}</div>
            </div>
          </div>
        )}
    </div>
  );
}
