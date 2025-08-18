#!/usr/bin/env python3
import os
import requests
import json
from pathlib import Path

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
project_name = "quiver-network"
docs_dir = "docs"

# API endpoint
url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"

# Headers
headers = {
    "Authorization": f"Bearer {api_token}"
}

# Collect all HTML and essential files
files = []
manifest = {}

print("Collecting files...")

for root, dirs, filenames in os.walk(docs_dir):
    # Skip hidden directories
    dirs[:] = [d for d in dirs if not d.startswith('.')]
    
    for filename in filenames:
        if filename.endswith(('.html', '.js', '.css', '.json', '.txt', '.md')):
            filepath = os.path.join(root, filename)
            relative_path = os.path.relpath(filepath, docs_dir)
            
            # Skip large files
            file_size = os.path.getsize(filepath)
            if file_size < 5 * 1024 * 1024:  # 5MB limit
                files.append(('file', (relative_path, open(filepath, 'rb'))))
                # For subdomain routing, ensure index files are at root
                if relative_path.endswith('/index.html'):
                    # Also add without the index.html suffix
                    dir_path = relative_path[:-11]  # Remove '/index.html'
                    if dir_path:
                        manifest[dir_path] = relative_path
                manifest[relative_path] = relative_path
                print(f"  Added: {relative_path} ({file_size // 1024}KB)")

print(f"\nTotal files to upload: {len(files)}")

# Create multipart data
multipart_data = [('manifest', (None, json.dumps(manifest)))] + files

print("\nUploading to Cloudflare Pages...")

# Make the request
try:
    response = requests.post(url, headers=headers, files=multipart_data, timeout=300)
    
    # Close all file handles
    for _, (_, file_handle) in files:
        file_handle.close()
    
    # Print result
    result = response.json()
    if result.get('success'):
        print("\n✅ Deployment successful!")
        deployment = result.get('result', {})
        print(f"ID: {deployment.get('id', 'N/A')}")
        print(f"URL: {deployment.get('url', 'N/A')}")
        print(f"Environment: {deployment.get('environment', 'N/A')}")
    else:
        print("\n❌ Deployment failed:")
        print(json.dumps(result, indent=2))
        
except Exception as e:
    print(f"\n❌ Error: {str(e)}")
    # Clean up file handles
    for _, (_, file_handle) in files:
        try:
            file_handle.close()
        except:
            pass