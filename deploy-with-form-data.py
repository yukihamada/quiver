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

# API endpoint for direct upload
url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"

headers = {
    "Authorization": f"Bearer {api_token}"
}

print("📦 Cloudflare Pages デプロイメント (Form-Data方式)\n")

# Create FormData with proper structure
files_to_upload = []
manifest = {}

# Walk through docs directory
for root, dirs, files in os.walk(docs_dir):
    # Skip hidden directories
    dirs[:] = [d for d in dirs if not d.startswith('.')]
    
    for filename in files:
        if not filename.startswith('.'):
            filepath = os.path.join(root, filename)
            relative_path = os.path.relpath(filepath, docs_dir)
            
            # Normalize path separators for web
            web_path = relative_path.replace('\\', '/')
            
            # Skip large files
            file_size = os.path.getsize(filepath)
            if file_size > 5 * 1024 * 1024:  # 5MB limit
                print(f"  スキップ: {web_path} ({file_size} bytes - too large)")
                continue
                
            # Read file content
            try:
                with open(filepath, 'rb') as f:
                    content = f.read()
                    
                # Add to files list
                files_to_upload.append(
                    ('file', (web_path, content, 'application/octet-stream'))
                )
                
                # Add to manifest
                manifest[web_path] = web_path
                
                print(f"  準備: {web_path} ({len(content)} bytes)")
                
            except Exception as e:
                print(f"  ❌ エラー: {filepath} - {e}")

print(f"\n合計ファイル数: {len(files_to_upload)}")

# Add manifest to files
files_to_upload.insert(0, ('manifest', (None, json.dumps(manifest), 'application/json')))

# Create the multipart request
print("\nアップロード中...")

try:
    response = requests.post(
        url,
        headers=headers,
        files=files_to_upload,
        timeout=300
    )
    
    result = response.json()
    
    if result.get('success'):
        deployment = result['result']
        print("\n✅ デプロイメント成功!")
        print(f"ID: {deployment['id']}")
        print(f"URL: {deployment['url']}")
        print(f"環境: {deployment['environment']}")
        
        # Wait a moment and test
        import time
        print("\n⏳ 10秒待機中...")
        time.sleep(10)
        
        # Test the deployment
        test_url = deployment['url']
        test_response = requests.get(test_url)
        print(f"\nテスト: {test_url}")
        print(f"ステータス: {test_response.status_code}")
        
        if test_response.status_code == 200:
            print("✅ デプロイメントが正常に動作しています！")
        else:
            print("⚠️  まだ404エラーです。Functions設定を確認してください。")
            
    else:
        print("\n❌ デプロイメント失敗:")
        print(json.dumps(result, indent=2))
        
except Exception as e:
    print(f"\n❌ エラー: {e}")

print("\n📝 次のステップ:")
print("1. Cloudflareダッシュボードでデプロイメントログを確認")
print("2. Functions & Routes設定を確認")
print("3. _worker.jsのエラーログを確認")