#!/usr/bin/env python3
import os
import requests
import json

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
project_name = "quiver-network-v2"
docs_dir = "docs"

headers = {
    "Authorization": f"Bearer {api_token}"
}

print("📦 完全デプロイメント for quiver-network-v2\n")

# Prepare all files
files_to_upload = []
manifest = {}

# Walk through docs directory
for root, dirs, files in os.walk(docs_dir):
    dirs[:] = [d for d in dirs if not d.startswith('.')]
    
    for filename in files:
        if not filename.startswith('.'):
            filepath = os.path.join(root, filename)
            relative_path = os.path.relpath(filepath, docs_dir).replace('\\', '/')
            
            # Skip large files
            file_size = os.path.getsize(filepath)
            if file_size > 10 * 1024 * 1024:  # 10MB limit
                print(f"  スキップ: {relative_path} ({file_size} bytes)")
                continue
            
            try:
                with open(filepath, 'rb') as f:
                    content = f.read()
                    files_to_upload.append(('file', (relative_path, content)))
                    manifest[relative_path] = relative_path
                    print(f"  準備: {relative_path}")
            except Exception as e:
                print(f"  エラー: {filepath} - {e}")

# Add manifest
files_to_upload.insert(0, ('manifest', (None, json.dumps(manifest))))

print(f"\n合計ファイル数: {len(files_to_upload) - 1}")
print("\nアップロード中...")

# Deploy
deploy_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"

try:
    response = requests.post(deploy_url, headers=headers, files=files_to_upload, timeout=300)
    result = response.json()
    
    if result.get('success'):
        deployment = result['result']
        print(f"\n✅ デプロイメント成功!")
        print(f"ID: {deployment['id']}")
        print(f"URL: {deployment['url']}")
        
        # Wait and test
        import time
        print("\n⏳ 30秒待機中...")
        time.sleep(30)
        
        print("\n📋 アクセステスト:")
        test_urls = [
            "https://quiver.network/",
            "https://www.quiver.network/",
            "https://api.quiver.network/",
            "https://docs.quiver.network/",
            "https://explorer.quiver.network/",
            "https://dashboard.quiver.network/"
        ]
        
        for url in test_urls:
            try:
                resp = requests.get(url, timeout=5)
                print(f"{url}: {resp.status_code}")
            except Exception as e:
                print(f"{url}: ❌ {str(e)}")
                
    else:
        print(f"\n❌ デプロイメント失敗: {result}")
        
except Exception as e:
    print(f"\n❌ エラー: {e}")

print("\n✅ 完了!")