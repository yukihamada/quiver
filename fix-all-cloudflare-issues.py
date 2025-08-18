#!/usr/bin/env python3
import requests
import json
import time

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
account_id = "08519319108846c5673d8dbf1a23f6a5"
zone_id = "a56354ca4082aa4640456f87304fde80"
project_name = "quiver-network"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🚀 Cloudflare完全自動修正スクリプト\n")

# Step 1: Fix root domain CNAME
print("1️⃣ ルートドメインのDNS修正...")
dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"

# Get root domain record
params = {"name": "quiver.network", "type": "CNAME"}
response = requests.get(dns_url, headers=headers, params=params).json()

if response.get('success') and response['result']:
    record = response['result'][0]
    record_id = record['id']
    current_content = record['content']
    
    print(f"現在のCNAME: {current_content}")
    
    if current_content != "quiver-network-dab.pages.dev":
        # Update CNAME
        update_data = {
            "type": "CNAME",
            "name": "quiver.network",
            "content": "quiver-network-dab.pages.dev",
            "proxied": True,
            "ttl": 1
        }
        
        update_url = f"{dns_url}/{record_id}"
        update_response = requests.patch(update_url, headers=headers, json=update_data)
        
        if update_response.status_code == 200:
            print("✅ ルートドメインのCNAMEを修正しました")
        else:
            print(f"❌ CNAME更新エラー: {update_response.json()}")
            
            # Try deleting and recreating
            print("CNAMEを削除して再作成を試みます...")
            delete_response = requests.delete(update_url, headers=headers)
            
            if delete_response.status_code == 200:
                print("✅ 古いCNAMEを削除しました")
                
                # Add CNAME flattening record
                create_response = requests.post(dns_url, headers=headers, json=update_data)
                if create_response.status_code == 200:
                    print("✅ 新しいCNAMEを作成しました")
                else:
                    print(f"❌ CNAME作成エラー: {create_response.json()}")

# Step 2: Enable CNAME flattening for root domain
print("\n2️⃣ CNAMEフラットニングを有効化...")
# This is automatically handled by Cloudflare for root domain CNAMEs

# Step 3: Update all subdomain CNAMEs
print("\n3️⃣ すべてのサブドメインCNAMEを確認・修正...")

subdomains = [
    "www", "api", "docs", "explorer", "dashboard", 
    "security", "quicpair", "playground", "status", 
    "blog", "community", "cdn"
]

for subdomain in subdomains:
    full_domain = f"{subdomain}.quiver.network"
    params = {"name": full_domain, "type": "CNAME"}
    response = requests.get(dns_url, headers=headers, params=params).json()
    
    if response.get('success') and response['result']:
        record = response['result'][0]
        if record['content'] != "quiver-network-dab.pages.dev":
            print(f"  修正中: {full_domain}")
            
            update_data = {
                "type": "CNAME",
                "name": full_domain,
                "content": "quiver-network-dab.pages.dev",
                "proxied": True,
                "ttl": 1
            }
            
            update_url = f"{dns_url}/{record['id']}"
            update_response = requests.patch(update_url, headers=headers, json=update_data)
            
            if update_response.status_code == 200:
                print(f"    ✅ 更新完了")
            else:
                print(f"    ❌ エラー: {update_response.json()}")
        else:
            print(f"  ✅ {full_domain} - 既に正しい設定")

# Step 4: Force SSL mode to Flexible temporarily
print("\n4️⃣ SSL設定を一時的にFlexibleに変更...")
ssl_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/settings/ssl"
ssl_response = requests.patch(ssl_url, headers=headers, json={"value": "flexible"})

if ssl_response.status_code == 200:
    print("✅ SSL設定をFlexibleに変更しました")
    
    # Wait and then set back to Full
    print("⏳ 10秒待機...")
    time.sleep(10)
    
    ssl_response = requests.patch(ssl_url, headers=headers, json={"value": "full"})
    if ssl_response.status_code == 200:
        print("✅ SSL設定をFullに戻しました")

# Step 5: Purge cache
print("\n5️⃣ キャッシュをパージ...")
purge_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache"
purge_data = {"purge_everything": True}

purge_response = requests.post(purge_url, headers=headers, json=purge_data)
if purge_response.status_code == 200:
    print("✅ キャッシュをパージしました")

# Step 6: Trigger custom domain verification
print("\n6️⃣ カスタムドメインの検証をトリガー...")

# Get Pages project custom domains
pages_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}"
pages_response = requests.get(pages_url, headers=headers).json()

if pages_response.get('success'):
    project = pages_response['result']
    print(f"プロジェクト: {project['name']}")
    print(f"カスタムドメイン数: {len(project.get('domains', []))}")
    
    # Try to re-add domains to trigger verification
    add_domain_url = f"{pages_url}/domains"
    
    for domain in ["quiver.network"] + [f"{sub}.quiver.network" for sub in subdomains]:
        print(f"\n  検証トリガー: {domain}")
        add_response = requests.post(add_domain_url, headers=headers, json={"name": domain})
        
        if add_response.status_code == 409:
            print(f"    ℹ️  既に追加済み - 検証待ち")
        elif add_response.status_code == 200:
            print(f"    ✅ 追加成功 - 検証開始")
        else:
            result = add_response.json()
            print(f"    ⚠️  {result.get('errors', [{}])[0].get('message', 'Unknown error')}")

# Step 7: Wait and test
print("\n7️⃣ DNS伝播を待機中（30秒）...")
time.sleep(30)

# Step 8: Test all domains
print("\n8️⃣ すべてのドメインをテスト...")

test_domains = ["quiver.network"] + [f"{sub}.quiver.network" for sub in subdomains]

success_count = 0
for domain in test_domains:
    url = f"https://{domain}/"
    try:
        response = requests.get(url, timeout=5, allow_redirects=True)
        status = response.status_code
        
        if status == 200:
            print(f"✅ {domain}: {status} - 成功！")
            success_count += 1
        elif status == 404:
            print(f"⚠️  {domain}: {status} - ファイルが見つかりません")
        elif status == 403:
            print(f"❌ {domain}: {status} - アクセス拒否（検証待ち）")
        else:
            print(f"❓ {domain}: {status}")
            
    except requests.exceptions.SSLError:
        print(f"🔒 {domain}: SSL証明書エラー（生成中）")
    except Exception as e:
        print(f"❌ {domain}: エラー - {str(e)}")

print(f"\n\n📊 結果: {success_count}/{len(test_domains)} ドメインが正常に動作")

if success_count < len(test_domains):
    print("\n⏳ まだ一部のドメインが検証中です。")
    print("💡 ヒント: 5-10分後に再度テストしてください。")
    print("\n🔄 再テストコマンド:")
    print("curl -I https://quiver.network/")
    print("curl -I https://docs.quiver.network/")
    print("curl -I https://api.quiver.network/")
else:
    print("\n🎉 すべてのドメインが正常に動作しています！")

# Step 9: Check Pages Functions logs
print("\n9️⃣ Pages Functionsのログを確認...")
try:
    # Try to get recent deployments
    deployments_url = f"https://api.cloudflare.com/client/v4/accounts/{account_id}/pages/projects/{project_name}/deployments"
    deployments = requests.get(deployments_url, headers=headers).json()
    
    if deployments.get('success') and deployments['result']:
        latest = deployments['result'][0]
        print(f"最新デプロイメント: {latest['id']}")
        print(f"ステータス: {latest['latest_stage']['status']}")
        
        if '_worker.js' in latest.get('files', {}):
            print("✅ _worker.js が検出されました")
        else:
            print("⚠️  _worker.js が見つかりません")
            
except Exception as e:
    print(f"ログ確認エラー: {e}")

print("\n✅ 自動修正スクリプト完了！")