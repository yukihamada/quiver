#!/usr/bin/env python3
import requests
import json

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
zone_id = "a56354ca4082aa4640456f87304fde80"
project_name = "quiver-network"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("📊 Cloudflare Pages カスタムドメイン詳細チェック\n")

# Get detailed project info including custom domains
url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}"
response = requests.get(url, headers=headers).json()

if not response.get('success'):
    print(f"❌ エラー: {response}")
    exit(1)

project = response['result']
domains = project.get('domains', [])

print(f"プロジェクト: {project['name']}")
print(f"サブドメイン: {project['subdomain']}")
print(f"プロダクションブランチ: {project.get('production_branch', 'N/A')}")
print(f"\nカスタムドメイン数: {len(domains)}")

# Get custom domain details (if API supports it)
print("\n📋 カスタムドメイン一覧:")
for i, domain in enumerate(domains, 1):
    print(f"\n{i}. {domain}")
    
    # Check if this is the Pages subdomain
    if domain.endswith('.pages.dev'):
        print(f"   タイプ: Cloudflare Pagesサブドメイン")
        print(f"   状態: ✅ 常にアクティブ")
    else:
        # Try to get domain status via different endpoint
        domain_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/custom_hostnames"
        params = {"hostname": domain}
        domain_response = requests.get(domain_url, headers=headers, params=params).json()
        
        if domain_response.get('success') and domain_response.get('result'):
            hostname = domain_response['result'][0]
            print(f"   状態: {hostname.get('status', 'unknown')}")
            print(f"   SSL: {hostname.get('ssl', {}).get('status', 'unknown')}")
        else:
            # Check DNS record
            dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"
            dns_params = {"name": domain}
            dns_response = requests.get(dns_url, headers=headers, params=dns_params).json()
            
            if dns_response.get('success') and dns_response.get('result'):
                record = dns_response['result'][0]
                print(f"   DNS: {record['type']} → {record['content']}")
                print(f"   プロキシ: {'✅' if record.get('proxied') else '❌'}")
                print(f"   TTL: {record.get('ttl', 'auto')}")

# Check latest deployment files
print("\n📁 最新デプロイメントのファイル構造をチェック:")
deployments_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"
deployments = requests.get(deployments_url, headers=headers).json()

if deployments.get('success') and deployments['result']:
    latest = deployments['result'][0]
    print(f"\nデプロイメントID: {latest['id']}")
    
    # Get deployment details
    deployment_detail_url = f"{deployments_url}/{latest['id']}"
    deployment_detail = requests.get(deployment_detail_url, headers=headers).json()
    
    if deployment_detail.get('success'):
        files = deployment_detail['result'].get('files', {})
        
        # Check key files
        key_files = [
            'index.html',
            'api/index.html', 
            'docs/index.html',
            'explorer/index.html',
            'dashboard/index.html',
            '_worker.js',
            '_routes.json',
            '_redirects'
        ]
        
        print("\n重要ファイルの存在確認:")
        for file in key_files:
            exists = file in files
            print(f"  {file}: {'✅' if exists else '❌'}")
        
        # Show first 10 files
        print(f"\nファイル総数: {len(files)}")
        print("最初の10ファイル:")
        for i, (path, _) in enumerate(list(files.items())[:10], 1):
            print(f"  {i}. {path}")

# Check if Pages Functions are enabled
print("\n⚙️ Pages Functions設定:")
deployment_configs = project.get('deployment_configs', {})
production_config = deployment_configs.get('production', {})

print(f"互換性日付: {production_config.get('compatibility_date', 'N/A')}")
print(f"ビルドコマンド: {production_config.get('build_config', {}).get('build_command', 'なし')}")
print(f"出力ディレクトリ: {production_config.get('build_config', {}).get('destination_dir', 'なし')}")
print(f"ルートディレクトリ: {production_config.get('build_config', {}).get('root_dir', 'なし')}")

# Final check - test actual endpoints
print("\n🌐 実際のアクセステスト:")
test_urls = [
    f"https://{project['subdomain']}/",
    f"https://{project['subdomain']}/index.html",
    f"https://{project['subdomain']}/api/index.html",
    "https://quiver.network/",
    "https://api.quiver.network/",
    "https://docs.quiver.network/"
]

for url in test_urls:
    try:
        response = requests.head(url, allow_redirects=True, timeout=5)
        print(f"{url}: {response.status_code}")
    except Exception as e:
        print(f"{url}: ❌ エラー - {str(e)}")

print("\n✅ チェック完了")