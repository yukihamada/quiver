#!/usr/bin/env python3
import requests
import json

# Configuration
api_token = "8sFAg2aVWcYm5rLZ7NHPJwtx_KswmzH9U3GOpC4n"
zone_id = "a56354ca4082aa4640456f87304fde80"

headers = {
    "Authorization": f"Bearer {api_token}",
    "Content-Type": "application/json"
}

print("🔧 ルートドメイン修正スクリプト\n")

# Step 1: Get current root domain DNS records
dns_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/dns_records"
params = {"name": "quiver.network"}

response = requests.get(dns_url, headers=headers, params=params).json()

if response.get('success'):
    records = response['result']
    print(f"現在のルートドメインレコード数: {len(records)}")
    
    for record in records:
        print(f"\nレコード ID: {record['id']}")
        print(f"タイプ: {record['type']}")
        print(f"コンテンツ: {record['content']}")
        print(f"プロキシ: {record.get('proxied', False)}")
        
        # If it's a CNAME, we need to delete it first
        if record['type'] == 'CNAME' and record['name'] == 'quiver.network':
            print("\n❌ ルートドメインにCNAMEは使用できません")
            print("削除しますか？ (実行する場合はスクリプトを編集してください)")
            
            # Uncomment to delete
            # delete_url = f"{dns_url}/{record['id']}"
            # delete_response = requests.delete(delete_url, headers=headers)
            # if delete_response.status_code == 200:
            #     print("✅ CNAMEレコードを削除しました")

# Step 2: Add A records for Cloudflare Pages
print("\n\n推奨される修正方法:")
print("1. Cloudflareダッシュボードでルートドメインの設定を確認")
print("2. CNAMEフラットニングを有効化")
print("3. または、以下のAレコードを追加:")
print("   - 172.67.215.147")
print("   - 104.21.16.195")

# Alternative: Try to update SSL settings
print("\n\nSSL設定を確認中...")
ssl_url = f"https://api.cloudflare.com/client/v4/zones/{zone_id}/settings/ssl"
ssl_response = requests.get(ssl_url, headers=headers).json()

if ssl_response.get('success'):
    current_ssl = ssl_response['result']['value']
    print(f"現在のSSL設定: {current_ssl}")
    
    if current_ssl != "flexible":
        print("SSL設定を'Flexible'に変更を試みます...")
        update_ssl = requests.patch(ssl_url, headers=headers, json={"value": "flexible"})
        if update_ssl.status_code == 200:
            print("✅ SSL設定を変更しました")
        else:
            print(f"⚠️  変更失敗: {update_ssl.json()}")

print("\n\n💡 最終手段:")
print("Cloudflareサポートに連絡して、.network TLDのカスタムドメイン設定を")
print("手動で有効化してもらう必要があるかもしれません。")