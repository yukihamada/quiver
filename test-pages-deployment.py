#!/usr/bin/env python3
import requests
import json

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
project_name = "quiver-network"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🧪 Cloudflare Pages デプロイメントテスト\n")

# Get deployment details
url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"
response = requests.get(url, headers=headers).json()

if response.get('success') and response['result']:
    latest = response['result'][0]
    deployment_id = latest['id']
    
    # Get full deployment details
    detail_url = f"{url}/{deployment_id}"
    detail = requests.get(detail_url, headers=headers).json()
    
    if detail.get('success'):
        deployment = detail['result']
        
        print(f"デプロイメントID: {deployment['id']}")
        print(f"URL: {deployment['url']}")
        print(f"エイリアス: {deployment.get('aliases', [])}")
        print(f"環境: {deployment['environment']}")
        print(f"ステージ: {deployment['latest_stage']['name']} - {deployment['latest_stage']['status']}")
        
        # Check build output
        build_config = deployment.get('build_config', {})
        print(f"\nビルド設定:")
        print(f"  コマンド: {build_config.get('build_command', 'なし')}")
        print(f"  出力ディレクトリ: {build_config.get('destination_dir', 'なし')}")
        print(f"  ルートディレクトリ: {build_config.get('root_dir', 'なし')}")
        
        # Check if functions are enabled
        if '_worker.js' in deployment.get('files', {}):
            print("\n✅ _worker.js が検出されました - Functions有効")
            
            # Check Functions logs (if available)
            functions_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments/{deployment_id}/functions"
            functions_response = requests.get(functions_url, headers=headers)
            
            if functions_response.status_code == 200:
                functions_data = functions_response.json()
                print("Functions情報:", functions_data)
        
        # Test direct file access
        print("\n📁 ファイルアクセステスト:")
        base_url = deployment['url']
        
        test_paths = [
            '',
            'index.html',
            'api/index.html',
            'docs/index.html',
            '_worker.js'
        ]
        
        for path in test_paths:
            test_url = f"{base_url}/{path}" if path else base_url
            try:
                resp = requests.get(test_url, timeout=5)
                content_type = resp.headers.get('content-type', 'unknown')
                print(f"  {path or '/'}: {resp.status_code} ({content_type})")
                
                # If 404, show first 200 chars of response
                if resp.status_code == 404:
                    print(f"    Response: {resp.text[:200]}...")
                elif resp.status_code == 200 and path == '':
                    print(f"    Title検索: {'<title>' in resp.text}")
                    
            except Exception as e:
                print(f"  {path or '/'}: ❌ {str(e)}")

# Check if we need to purge cache
print("\n🔄 キャッシュのパージを試みます...")
zone_id = "a56354ca4082aa4640456f87304fde80"
purge_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache"
purge_data = {
    "purge_everything": True
}

purge_response = requests.post(purge_url, headers=headers, json=purge_data)
if purge_response.status_code == 200:
    print("✅ キャッシュをパージしました")
else:
    print(f"⚠️  キャッシュパージエラー: {purge_response.status_code}")

print("\n💡 推奨事項:")
print("1. Cloudflare Pagesの_worker.jsが正しく動作していない可能性があります")
print("2. ルートディレクトリ設定を確認してください")
print("3. Functions logsでエラーを確認してください")