# MaizAI Conversation Frontend

A Next.js-based conversation interface for the MaizAI API, featuring configurable AI models, conversation management, and comprehensive error reporting.

## Features

- 🤖 **AI Conversation Interface**: Chat with AI models through a clean, intuitive interface
- ⚙️ **Configurable Parameters**: Adjust model, provider, temperature, max tokens, and more
- 🌐 **Flexible API Configuration**: Connect to any MaizAI API server
- 📝 **Context Management**: Automatic conversation context persistence
- 🚨 **Error Reporting**: Detailed error messages with troubleshooting guidance
- 📱 **Responsive Design**: Works on desktop and mobile devices

## Quick Start

### 1. Configure API Connection

Copy the example environment file and configure your API server:

```bash
cp .env.example .env.local
```

Edit `.env.local`:
```bash
# Set your MaizAI API server URL
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

### 2. Install Dependencies

```bash
npm install
```

### 3. Start Development Server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the application.

## Configuration

### API Base URL

The application can be configured to connect to different MaizAI API servers:

#### Method 1: Environment Variables (Recommended)
```bash
# .env.local
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

#### Method 2: UI Configuration
1. Open the application
2. Edit the "API Base URL" field in the Configuration panel
3. Enter your API server URL (e.g., `http://localhost:8080`)

### Configuration Options

The application supports configuring:
- **API Base URL**: The MaizAI API server endpoint
- **Model**: AI model to use (e.g., `gpt-4`, `claude-3`)
- **Provider**: AI provider (e.g., `openai`, `anthropic`)
- **Max Tokens**: Maximum tokens for responses
- **Temperature**: Response creativity (0.0 to 2.0)
- **System Prompt**: Optional system instructions
- **Context Name**: Name for new conversations

### Detailed Configuration Guide

For comprehensive configuration instructions, see [API_CONFIGURATION.md](./API_CONFIGURATION.md).

## Usage

1. **Configure your API server** (see Configuration section above)
2. **Open the application** in your browser
3. **Adjust settings** in the Configuration panel if needed
4. **Start a conversation** by typing a message and clicking Send
5. **Continue the conversation** - context is automatically maintained
6. **Clear conversation** using the Clear button to start fresh

## Error Handling

The application provides detailed error reporting:
- **Error messages** appear directly in the chat
- **Error banners** show detailed debugging information
- **Copy Error Report** button for easy troubleshooting
- **Built-in troubleshooting** tips for common issues

## Building for Production

```bash
# Build the application
npm run build

# Start production server
npm start
```

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
