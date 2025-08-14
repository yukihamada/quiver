# iOS Setup Instructions

## Using Shortcuts App

1. Open Shortcuts app on iPhone
2. Create new shortcut
3. Add action: "Get Contents of URL"
4. Set URL: http://192.168.0.194:8080/generate
5. Set Method: POST
6. Add Headers:
   - Content-Type: application/json
7. Set Body: JSON with format:
   ```json
   {
     "prompt": "Your question here"
   }
   ```
8. Add action: "Get Dictionary from Input"
9. Add action: "Get Text from Input" (for completion field)
10. Save and run!

## Using HTTP Client Apps

Recommended apps:
- HTTP Bot
- RESTed
- Paw

Configure with:
- URL: http://192.168.0.194:8080/generate
- Method: POST
- Headers: Content-Type: application/json
- Body: {"prompt": "Your question"}
