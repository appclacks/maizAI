# API Configuration Guide

This guide explains how to configure the MaizAI API base URL for the conversation application.

## Configuration Methods

### 1. Environment Variables (Recommended)

The easiest way to configure the API base URL is using environment variables:

#### Development Environment

1. **Create or edit `.env.local`** in the project root:
   ```bash
   # Example for local development
   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
   
   # Example for production
   NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com
   ```

2. **Restart the development server**:
   ```bash
   npm run dev
   ```

#### Production Environment

For production deployments, set the environment variable in your hosting platform:

**Vercel:**
```bash
vercel env add NEXT_PUBLIC_API_BASE_URL
```

**Netlify:**
- Go to Site settings → Environment variables
- Add: `NEXT_PUBLIC_API_BASE_URL` = `https://your-api-domain.com`

**Docker:**
```dockerfile
ENV NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com
```

**Docker Compose:**
```yaml
environment:
  - NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com
```

### 2. Runtime Configuration (UI)

You can also configure the API URL directly in the application interface:

1. **Open the application** in your browser
2. **Find the Configuration Panel** at the top of the page
3. **Edit the "API Base URL" field**
4. **Enter your API server URL** (e.g., `http://localhost:8080`)
5. **Start a conversation** to test the connection

## URL Format Guidelines

### ✅ Correct Format
```
http://localhost:8080
https://api.maizai.com
https://your-domain.com:3000
```

### ❌ Incorrect Format
```
http://localhost:8080/          (trailing slash)
http://localhost:8080/api/      (includes path)
localhost:8080                  (missing protocol)
```

## Common Configuration Examples

### Local Development
```bash
# Standard local development
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

# Custom port
NEXT_PUBLIC_API_BASE_URL=http://localhost:3001

# Local IP for testing from other devices
NEXT_PUBLIC_API_BASE_URL=http://192.168.1.100:8080
```

### Production
```bash
# HTTPS production domain
NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com

# Custom subdomain
NEXT_PUBLIC_API_BASE_URL=https://maizai-api.yourdomain.com

# Non-standard port
NEXT_PUBLIC_API_BASE_URL=https://your-domain.com:8443
```

## Testing Your Configuration

### 1. Check Environment Variables
```bash
# View current environment variables
echo $NEXT_PUBLIC_API_BASE_URL

# Or check in the browser console
console.log(process.env.NEXT_PUBLIC_API_BASE_URL)
```

### 2. Test API Endpoint
```bash
# Test if your API server is accessible
curl -X POST http://localhost:8080/api/v1/conversation \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"test"}]}'
```

### 3. Check Network Tab
1. Open browser **Developer Tools** (F12)
2. Go to **Network** tab
3. Send a message in the application
4. Look for the API request and verify the URL is correct

## Troubleshooting

### Common Issues

**1. CORS Errors**
- Ensure your API server allows requests from your frontend domain
- Check that CORS headers are properly configured on the server

**2. Connection Refused**
- Verify the API server is running
- Check if the port is correct
- Ensure no firewall is blocking the connection

**3. 404 Not Found**
- Verify the API base URL doesn't include `/api/v1/conversation`
- Check that the API server has the correct endpoint

**4. Environment Variables Not Working**
- Ensure the variable starts with `NEXT_PUBLIC_`
- Restart the development server after changes
- Check there are no typos in the variable name

### Error Reporting

The application includes detailed error reporting:
- **Error messages** appear in the chat
- **Error banner** shows detailed information
- **Copy Error Report** button includes configuration details
- **Troubleshooting tips** are provided automatically

## Security Considerations

### Development
- Use `http://localhost` for local development
- Never commit sensitive URLs to version control

### Production
- Always use HTTPS in production
- Use environment variables for sensitive configuration
- Implement proper authentication and authorization
- Consider using API keys or tokens for authentication

## Examples by Deployment Platform

### Vercel
```bash
# Set environment variable
vercel env add NEXT_PUBLIC_API_BASE_URL

# Or in vercel.json
{
  "env": {
    "NEXT_PUBLIC_API_BASE_URL": "https://api.maizai.com"
  }
}
```

### Netlify
```bash
# netlify.toml
[build.environment]
  NEXT_PUBLIC_API_BASE_URL = "https://api.maizai.com"
```

### Railway
```bash
# Set via Railway CLI
railway variables set NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com
```

### Self-hosted
```bash
# Export environment variable
export NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com

# Or add to .bashrc/.zshrc
echo 'export NEXT_PUBLIC_API_BASE_URL=https://api.maizai.com' >> ~/.bashrc
```

## Quick Start

1. **Copy the example environment file**:
   ```bash
   cp .env.example .env.local
   ```

2. **Edit the API URL**:
   ```bash
   # Edit .env.local
   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
   ```

3. **Start the development server**:
   ```bash
   npm run dev
   ```

4. **Test the connection** by sending a message in the application

That's it! Your application is now configured to connect to your MaizAI API server.