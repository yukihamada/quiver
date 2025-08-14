# QUIVer iOS Quick Setup Guide

## 🚀 Service is Running!

Your QUIVer inference service is now accessible at:
- **Local Mac**: http://localhost:8080/generate
- **iPhone/iPad**: http://192.168.0.194:8080/generate

## 📱 Method 1: Using iOS Shortcuts (Recommended)

1. **Download this shortcut**: [QUIVer AI Assistant](https://www.icloud.com/shortcuts/example)
   
   Or create manually:
   
2. Open **Shortcuts** app
3. Tap **+** to create new shortcut
4. Add these actions:

   a. **Text** action:
      - Enter your question

   b. **Get Contents of URL** action:
      - URL: `http://192.168.0.194:8080/generate`
      - Method: `POST`
      - Headers: 
        - `Content-Type`: `application/json`
      - Request Body: `JSON`
      - JSON: 
        ```json
        {
          "prompt": "Text from previous action"
        }
        ```

   c. **Get Dictionary Value** action:
      - Get: `Value for completion`
      - From: `Contents of URL from previous action`

   d. **Show Result** action:
      - Text: `Dictionary Value from previous action`

5. Name it "Ask AI" and add to Home Screen

## 📱 Method 2: Using a-Shell app (Terminal)

1. Install **a-Shell** from App Store
2. Run:
   ```bash
   curl -X POST http://192.168.0.194:8080/generate \
     -H "Content-Type: application/json" \
     -d '{"prompt": "Your question here"}'
   ```

## 📱 Method 3: Using HTTP Bot app

1. Install **HTTP Bot** from App Store
2. Create new request:
   - URL: `http://192.168.0.194:8080/generate`
   - Method: `POST`
   - Headers: `Content-Type: application/json`
   - Body:
     ```json
     {
       "prompt": "What is the meaning of life?"
     }
     ```

## 🧪 Test Examples

Try these prompts:
- "What is the capital of Japan?"
- "Explain quantum computing in simple terms"
- "Write a haiku about programming"
- "Calculate 15% tip on $84"

## 🔧 Troubleshooting

1. **Connection refused**: Make sure your iPhone is on the same WiFi network
2. **Timeout**: The first request may take longer as the model loads
3. **No response**: Check if the gateway is still running on your Mac

## 📊 Service Status

Check if service is healthy:
```
http://192.168.0.194:8080/health
```

## 🛑 Stopping the Service

On your Mac terminal, press `Ctrl+C` or run:
```bash
pkill gateway_simple
```