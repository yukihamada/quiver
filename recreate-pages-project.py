#!/usr/bin/env python3
import requests
import json
import time
import os

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
zone_id = "a56354ca4082aa4640456f87304fde80"
old_project_name = "quiver-network"
new_project_name = "quiver-network-v2"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🔄 Cloudflare Pagesプロジェクト再作成スクリプト\n")

# Step 1: Delete old project
print("1️⃣ 古いプロジェクトを削除...")
delete_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{old_project_name}"
delete_response = requests.delete(delete_url, headers=headers)

if delete_response.status_code == 200:
    print("✅ 古いプロジェクトを削除しました")
    time.sleep(5)
elif delete_response.status_code == 404:
    print("ℹ️  プロジェクトが既に存在しません")
else:
    print(f"⚠️  削除エラー: {delete_response.json()}")
    print("続行しますか？ (Ctrl+Cで中止)")
    time.sleep(3)

# Step 2: Create new project
print(f"\n2️⃣ 新しいプロジェクト '{new_project_name}' を作成...")
create_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects"
create_data = {
    "name": new_project_name,
    "production_branch": "main",
    "deployment_configs": {
        "production": {
            "compatibility_date": "2024-01-01",
            "compatibility_flags": [],
            "build_config": {
                "build_command": "",
                "destination_dir": "",
                "root_dir": ""
            }
        }
    }
}

create_response = requests.post(create_url, headers=headers, json=create_data)

if create_response.status_code in [200, 201]:
    project = create_response.json()['result']
    print(f"✅ プロジェクト作成成功!")
    print(f"   名前: {project['name']}")
    print(f"   サブドメイン: {project['subdomain']}")
    new_subdomain = project['subdomain']
else:
    print(f"❌ プロジェクト作成エラー: {create_response.json()}")
    exit(1)

# Step 3: Deploy initial files
print("\n3️⃣ 初期ファイルをデプロイ...")

# Prepare files
deploy_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{new_project_name}/deployments"

# Read actual files from docs directory
files_to_upload = []
manifest = {}

docs_dir = "docs"
important_files = [
    "index.html",
    "api/index.html",
    "docs/index.html",
    "explorer/index.html",
    "dashboard/index.html",
    "security/index.html",
    "quicpair/index.html",
    "playground/index.html",
    "status/index.html",
    "blog/index.html",
    "community/index.html"
]

# Add important files first
for file_path in important_files:
    full_path = os.path.join(docs_dir, file_path)
    if os.path.exists(full_path):
        with open(full_path, 'rb') as f:
            content = f.read()
            files_to_upload.append(('file', (file_path, content, 'text/html')))
            manifest[file_path] = file_path
            print(f"  追加: {file_path}")

# Add manifest
files_to_upload.insert(0, ('manifest', (None, json.dumps(manifest), 'application/json')))

print("\nデプロイ中...")
deploy_response = requests.post(
    deploy_url,
    headers={"Authorization": f"Bearer {api_token}"},
    files=files_to_upload,
    timeout=60
)

if deploy_response.status_code == 200:
    deployment = deploy_response.json()['result']
    print(f"✅ デプロイメント成功!")
    print(f"   ID: {deployment['id']}")
    print(f"   URL: {deployment['url']}")
else:
    print(f"❌ デプロイメントエラー: {deploy_response.json()}")

# Step 4: Update DNS records
print(f"\n4️⃣ DNSレコードを新しいサブドメイン '{new_subdomain}' に更新...")

dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"

# Get all CNAME records pointing to old subdomain
params = {"type": "CNAME"}
response = requests.get(dns_url, headers=headers, params=params).json()

if response.get('success'):
    for record in response['result']:
        if 'quiver-network' in record['content'] and record['content'].endswith('.pages.dev'):
            print(f"  更新: {record['name']} → {new_subdomain}")
            
            update_data = {
                "type": "CNAME",
                "name": record['name'],
                "content": new_subdomain,
                "proxied": True,
                "ttl": 1
            }
            
            update_url = f"{dns_url}/{record['id']}"
            update_response = requests.patch(update_url, headers=headers, json=update_data)
            
            if update_response.status_code == 200:
                print(f"    ✅ 更新完了")
            else:
                print(f"    ❌ エラー: {update_response.json()}")

# Step 5: Add custom domains to new project
print(f"\n5️⃣ カスタムドメインを新しいプロジェクトに追加...")

domains = [
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

add_domain_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{new_project_name}/domains"

for domain in domains:
    print(f"  追加: {domain}")
    add_response = requests.post(add_domain_url, headers=headers, json={"name": domain})
    
    if add_response.status_code in [200, 201]:
        print(f"    ✅ 追加成功")
    elif add_response.status_code == 409:
        print(f"    ℹ️  既に存在")
    else:
        result = add_response.json()
        print(f"    ⚠️  {result.get('errors', [{}])[0].get('message', 'Unknown error')}")

# Step 6: Wait and test
print("\n6️⃣ DNS伝播を待機中（30秒）...")
time.sleep(30)

# Step 7: Final test
print("\n7️⃣ 最終テスト...")

test_urls = [
    f"https://{new_subdomain}/",
    "https://quiver.network/",
    "https://api.quiver.network/",
    "https://docs.quiver.network/"
]

success_count = 0
for url in test_urls:
    try:
        response = requests.get(url, timeout=5, allow_redirects=True)
        print(f"{url}: {response.status_code}")
        if response.status_code == 200:
            success_count += 1
    except Exception as e:
        print(f"{url}: ❌ {str(e)}")

print(f"\n\n✅ プロジェクト再作成完了!")
print(f"成功率: {success_count}/{len(test_urls)}")

if success_count < len(test_urls):
    print("\n⏳ まだDNSが伝播中の可能性があります。")
    print("5-10分後に再度確認してください。")

print(f"\n📝 新しいプロジェクト情報:")
print(f"プロジェクト名: {new_project_name}")
print(f"サブドメイン: {new_subdomain}")