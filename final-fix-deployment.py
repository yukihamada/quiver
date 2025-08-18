#!/usr/bin/env python3
import os
import requests
import json
import base64

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
project_name = "quiver-network"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🔧 最終修正: ファイル構造の問題を解決\n")

# Check current deployment structure
print("1️⃣ 現在のデプロイメント構造を確認...")
deployments_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"
response = requests.get(deployments_url, headers=headers).json()

if response.get('success') and response['result']:
    latest = response['result'][0]
    deployment_id = latest['id']
    
    # Get detailed deployment info
    detail_url = f"{deployments_url}/{deployment_id}"
    detail = requests.get(detail_url, headers=headers).json()
    
    if detail.get('success'):
        files = detail['result'].get('files', {})
        print(f"ファイル総数: {len(files)}")
        
        # Check file paths
        has_root_index = 'index.html' in files
        has_worker = '_worker.js' in files
        
        print(f"ルートindex.html: {'✅' if has_root_index else '❌'}")
        print(f"_worker.js: {'✅' if has_worker else '❌'}")
        
        # Check subdomain files
        subdomain_files = {
            'api': 'api/index.html' in files,
            'docs': 'docs/index.html' in files,
            'explorer': 'explorer/index.html' in files,
            'dashboard': 'dashboard/index.html' in files,
        }
        
        print("\nサブドメインファイル:")
        for sub, exists in subdomain_files.items():
            print(f"  {sub}/index.html: {'✅' if exists else '❌'}")

# Create a simple test deployment
print("\n2️⃣ シンプルなテストデプロイメントを作成...")

# Create test files
test_files = {
    "index.html": """<!DOCTYPE html>
<html>
<head>
    <title>QUIVer Network</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>QUIVer Network - Root</h1>
    <p>ルートドメインのテストページ</p>
    <ul>
        <li><a href="https://api.quiver.network/">API</a></li>
        <li><a href="https://docs.quiver.network/">Docs</a></li>
        <li><a href="https://explorer.quiver.network/">Explorer</a></li>
    </ul>
</body>
</html>""",
    
    "api/index.html": """<!DOCTYPE html>
<html>
<head>
    <title>QUIVer API</title>
</head>
<body>
    <h1>API Subdomain Test</h1>
    <p>APIサブドメインのテストページ</p>
</body>
</html>""",
    
    "docs/index.html": """<!DOCTYPE html>
<html>
<head>
    <title>QUIVer Docs</title>
</head>
<body>
    <h1>Docs Subdomain Test</h1>
    <p>Docsサブドメインのテストページ</p>
</body>
</html>""",
    
    "_routes.json": json.dumps({
        "version": 1,
        "include": ["/*"],
        "exclude": []
    })
}

# Deploy test files
deploy_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"

# Prepare multipart data
files_data = []
manifest = {}

for path, content in test_files.items():
    files_data.append(('file', (path, content.encode('utf-8'), 'text/html')))
    manifest[path] = path

# Add manifest
files_data.insert(0, ('manifest', (None, json.dumps(manifest), 'application/json')))

print("テストファイルをデプロイ中...")

try:
    response = requests.post(
        deploy_url,
        headers={"Authorization": f"Bearer {api_token}"},
        files=files_data,
        timeout=30
    )
    
    result = response.json()
    
    if result.get('success'):
        deployment = result['result']
        print(f"\n✅ テストデプロイメント成功!")
        print(f"ID: {deployment['id']}")
        print(f"URL: {deployment['url']}")
        
        # Test the deployment
        import time
        print("\n⏳ 15秒待機中...")
        time.sleep(15)
        
        # Test URLs
        print("\n3️⃣ デプロイメントをテスト...")
        test_urls = [
            deployment['url'],
            "https://quiver.network/",
            "https://api.quiver.network/",
            "https://docs.quiver.network/"
        ]
        
        for url in test_urls:
            try:
                resp = requests.get(url, timeout=5)
                print(f"{url}: {resp.status_code}")
                if resp.status_code == 200 and "テストページ" in resp.text:
                    print("  ✅ テストコンテンツが表示されています!")
            except Exception as e:
                print(f"{url}: ❌ {str(e)}")
                
    else:
        print(f"\n❌ デプロイメント失敗: {result}")
        
except Exception as e:
    print(f"\n❌ エラー: {e}")

print("\n4️⃣ 問題の診断...")
print("\n考えられる原因:")
print("1. _worker.jsが正しく動作していない")
print("2. ファイルパスのマッピングに問題がある")
print("3. Cloudflare Pagesの設定に問題がある")

print("\n💡 解決策:")
print("1. プロジェクトを削除して再作成")
print("2. または、Cloudflareサポートに連絡")
print("\n実行コマンド:")
print("# プロジェクト削除")
print(f"curl -X DELETE 'https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}' \\")
print(f"  -H 'Authorization: Bearer {api_token}'")
print("\n# 新規プロジェクト作成")
print("wrangler pages project create quiver-network --production-branch=main")