#!/usr/bin/env python3
import os
import requests
import json
import time

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
zone_id = "a56354ca4082aa4640456f87304fde80"  # quiver.network zone ID
project_name = "quiver-network"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🔧 Cloudflare Pages修正スクリプト開始...\n")

# Step 1: Get current project info
print("1️⃣ プロジェクト情報を取得中...")
url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}"
response = requests.get(url, headers=headers)
project = response.json()

if not project.get('success'):
    print(f"❌ プロジェクト取得エラー: {project}")
    exit(1)

print(f"✅ プロジェクト名: {project['result']['name']}")
print(f"✅ サブドメイン: {project['result']['subdomain']}")

# Step 2: Get custom domains status
print("\n2️⃣ カスタムドメインの状態を確認中...")
domains = project['result'].get('domains', [])

if domains:
    print(f"カスタムドメイン数: {len(domains)}")
    for domain in domains[:5]:  # Show first 5
        print(f"  - {domain}")
else:
    print("❌ カスタムドメインが設定されていません")

# Step 3: Check latest deployment
print("\n3️⃣ 最新のデプロイメントを確認中...")
deployments_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"
deployments = requests.get(deployments_url, headers=headers).json()

if deployments.get('success') and deployments['result']:
    latest = deployments['result'][0]
    print(f"✅ デプロイメントID: {latest['id']}")
    print(f"✅ URL: {latest['url']}")
    print(f"✅ ステータス: {latest['latest_stage']['status']}")
    
    # Check if deployment has files
    if 'files' in latest:
        file_count = len(latest.get('files', {}))
        print(f"✅ ファイル数: {file_count}")
else:
    print("❌ デプロイメントが見つかりません")

# Step 4: Update Pages project settings (if needed)
print("\n4️⃣ プロジェクト設定を更新中...")

# Update build configuration
update_data = {
    "deployment_configs": {
        "production": {
            "compatibility_date": "2024-01-01",
            "build_config": {
                "build_command": "",
                "destination_dir": "/",
                "root_dir": ""
            }
        }
    }
}

update_response = requests.patch(url, headers=headers, json=update_data)
if update_response.status_code == 200:
    print("✅ プロジェクト設定を更新しました")
else:
    print(f"⚠️  設定更新エラー: {update_response.status_code}")

# Step 5: Add custom domains via API (if possible)
print("\n5️⃣ カスタムドメインを追加中...")

custom_domains = [
    "quiver.network",
    "www.quiver.network",
    "api.quiver.network",
    "docs.quiver.network",
    "explorer.quiver.network",
    "dashboard.quiver.network",
    "security.quiver.network",
    "quicpair.quiver.network",
    "playground.quiver.network",
    "status.quiver.network",
    "blog.quiver.network",
    "community.quiver.network",
    "cdn.quiver.network"
]

# Try to add custom domains
for domain in custom_domains:
    print(f"\n追加中: {domain}")
    
    # First, check if DNS record exists
    dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"
    dns_params = {"name": domain, "type": "CNAME"}
    dns_records = requests.get(dns_url, headers=headers, params=dns_params).json()
    
    if dns_records.get('success') and dns_records['result']:
        record = dns_records['result'][0]
        print(f"  DNS: {record['type']} → {record['content']}")
        
        # Try to add custom domain to Pages
        add_domain_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/domains"
        add_data = {"name": domain}
        
        add_response = requests.post(add_domain_url, headers=headers, json=add_data)
        
        if add_response.status_code == 200:
            print(f"  ✅ カスタムドメイン追加成功")
        elif add_response.status_code == 409:
            print(f"  ℹ️  既に追加済み")
        else:
            result = add_response.json()
            error_msg = result.get('errors', [{}])[0].get('message', 'Unknown error')
            print(f"  ⚠️  エラー: {error_msg}")
    else:
        print(f"  ❌ DNSレコードが見つかりません")

# Step 6: Force SSL certificate generation
print("\n6️⃣ SSL証明書の生成を確認中...")

# Get zone SSL settings
ssl_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/settings/ssl"
ssl_response = requests.get(ssl_url, headers=headers).json()

if ssl_response.get('success'):
    ssl_mode = ssl_response['result']['value']
    print(f"SSL設定: {ssl_mode}")
    
    if ssl_mode != "full":
        # Update to Full SSL
        update_ssl = requests.patch(ssl_url, headers=headers, json={"value": "full"})
        if update_ssl.status_code == 200:
            print("✅ SSL設定を'Full'に更新しました")

# Step 7: Alternative approach - Update DNS records to use A record for root
print("\n7️⃣ ルートドメインのDNS設定を確認中...")

# Get root domain record
root_dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"
root_params = {"name": "quiver.network", "type": "A"}
root_records = requests.get(root_dns_url, headers=headers, params=root_params).json()

if root_records.get('success') and not root_records['result']:
    print("ℹ️  ルートドメインのAレコードがありません")
    print("   Cloudflare PagesはルートドメインにCNAMEを使用できません")
    
    # For root domain, we need to use Cloudflare's special CNAME flattening
    # or add an A record pointing to Cloudflare Pages IPs
    pages_ips = ["172.67.215.147", "104.21.16.195"]  # Cloudflare Pages IPs
    
    for ip in pages_ips:
        print(f"   Aレコード追加中: {ip}")
        add_a_record = {
            "type": "A",
            "name": "@",
            "content": ip,
            "proxied": True,
            "ttl": 1
        }
        
        add_response = requests.post(root_dns_url, headers=headers, json=add_a_record)
        if add_response.status_code == 200:
            print(f"   ✅ Aレコード追加成功")
        else:
            print(f"   ⚠️  エラー: {add_response.json()}")

print("\n✅ 修正スクリプト完了")
print("\n📝 次のステップ:")
print("1. 5-10分待ってDNSが伝播するのを待つ")
print("2. 各サブドメインにアクセスして確認")
print("3. まだ403/404エラーが出る場合は、Cloudflareダッシュボードで手動確認が必要")
print("\n💡 ヒント: .network TLDの制限により、一部の操作はAPIで実行できない場合があります")