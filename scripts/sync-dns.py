#!/usr/bin/env python3
"""
Sync DNS records from dns/records.yaml to Cloudflare
"""

import os
import sys
import yaml
import json
import requests
from typing import Dict, List, Any

# Cloudflare API configuration
CF_API_TOKEN = os.environ.get('CLOUDFLARE_API_TOKEN')
CF_ZONE_ID = os.environ.get('CLOUDFLARE_ZONE_ID')
CF_API_BASE = 'https://api.cloudflare.com/client/v4'

def load_dns_config(file_path: str = 'dns/records.yaml') -> Dict[str, Any]:
    """Load DNS configuration from YAML file"""
    with open(file_path, 'r') as f:
        return yaml.safe_load(f)

def get_existing_records() -> List[Dict[str, Any]]:
    """Fetch existing DNS records from Cloudflare"""
    headers = {
        'Authorization': f'Bearer {CF_API_TOKEN}',
        'Content-Type': 'application/json'
    }
    
    all_records = []
    page = 1
    
    while True:
        response = requests.get(
            f'{CF_API_BASE}/zones/{CF_ZONE_ID}/dns_records',
            headers=headers,
            params={'page': page, 'per_page': 100}
        )
        
        if response.status_code != 200:
            print(f"Error fetching DNS records: {response.text}")
            sys.exit(1)
        
        data = response.json()
        all_records.extend(data['result'])
        
        if page >= data['result_info']['total_pages']:
            break
        page += 1
    
    return all_records

def create_record(record: Dict[str, Any]) -> bool:
    """Create a new DNS record"""
    headers = {
        'Authorization': f'Bearer {CF_API_TOKEN}',
        'Content-Type': 'application/json'
    }
    
    # Prepare record data
    data = {
        'type': record['type'],
        'name': record['name'],
        'content': record['content'],
        'ttl': 1 if record.get('ttl') == 'auto' else record.get('ttl', 1),
        'proxied': record.get('proxied', True)
    }
    
    if record.get('comment'):
        data['comment'] = record['comment']
    
    # Add SRV-specific fields
    if record['type'] == 'SRV':
        data['data'] = {
            'priority': record.get('priority', 10),
            'weight': record.get('weight', 10),
            'port': record['port'],
            'target': record['target']
        }
        del data['content']
    
    response = requests.post(
        f'{CF_API_BASE}/zones/{CF_ZONE_ID}/dns_records',
        headers=headers,
        json=data
    )
    
    if response.status_code == 200:
        print(f"✅ Created: {record['name']} ({record['type']})")
        return True
    else:
        print(f"❌ Failed to create {record['name']}: {response.text}")
        return False

def update_record(record_id: str, record: Dict[str, Any]) -> bool:
    """Update an existing DNS record"""
    headers = {
        'Authorization': f'Bearer {CF_API_TOKEN}',
        'Content-Type': 'application/json'
    }
    
    # Prepare record data
    data = {
        'type': record['type'],
        'name': record['name'],
        'content': record['content'],
        'ttl': 1 if record.get('ttl') == 'auto' else record.get('ttl', 1),
        'proxied': record.get('proxied', True)
    }
    
    if record.get('comment'):
        data['comment'] = record['comment']
    
    # Add SRV-specific fields
    if record['type'] == 'SRV':
        data['data'] = {
            'priority': record.get('priority', 10),
            'weight': record.get('weight', 10),
            'port': record['port'],
            'target': record['target']
        }
        del data['content']
    
    response = requests.put(
        f'{CF_API_BASE}/zones/{CF_ZONE_ID}/dns_records/{record_id}',
        headers=headers,
        json=data
    )
    
    if response.status_code == 200:
        print(f"✅ Updated: {record['name']} ({record['type']})")
        return True
    else:
        print(f"❌ Failed to update {record['name']}: {response.text}")
        return False

def delete_record(record_id: str, name: str, record_type: str) -> bool:
    """Delete a DNS record"""
    headers = {
        'Authorization': f'Bearer {CF_API_TOKEN}',
        'Content-Type': 'application/json'
    }
    
    response = requests.delete(
        f'{CF_API_BASE}/zones/{CF_ZONE_ID}/dns_records/{record_id}',
        headers=headers
    )
    
    if response.status_code == 200:
        print(f"🗑️  Deleted: {name} ({record_type})")
        return True
    else:
        print(f"❌ Failed to delete {name}: {response.text}")
        return False

def sync_dns_records():
    """Main sync function"""
    print("🔄 Syncing DNS records to Cloudflare...")
    
    # Load configuration
    config = load_dns_config()
    domain = config['domain']
    
    # Get all records from config
    all_config_records = []
    for section in ['frontend', 'backend', 'bootstrap', 'srv']:
        if section in config and config[section]:
            all_config_records.extend(config[section])
    
    # Get existing records from Cloudflare
    existing_records = get_existing_records()
    existing_map = {}
    
    for record in existing_records:
        # Skip non-managed records
        if not record['name'].endswith(domain):
            continue
        
        key = f"{record['name']}:{record['type']}"
        existing_map[key] = record
    
    # Track processed records
    processed = set()
    stats = {'created': 0, 'updated': 0, 'deleted': 0, 'unchanged': 0}
    
    # Process config records
    for record in all_config_records:
        # Handle @ for root domain
        if record['name'] == '@':
            full_name = domain
        else:
            full_name = f"{record['name']}.{domain}"
        
        key = f"{full_name}:{record['type']}"
        processed.add(key)
        
        if key in existing_map:
            # Check if update needed
            existing = existing_map[key]
            needs_update = (
                existing['content'] != record['content'] or
                existing['proxied'] != record.get('proxied', True)
            )
            
            if needs_update:
                if update_record(existing['id'], {**record, 'name': full_name}):
                    stats['updated'] += 1
            else:
                print(f"✓ Unchanged: {record['name']} ({record['type']})")
                stats['unchanged'] += 1
        else:
            # Create new record
            if create_record({**record, 'name': full_name}):
                stats['created'] += 1
    
    # Delete records not in config
    for key, record in existing_map.items():
        if key not in processed:
            # Skip certain record types
            if record['type'] in ['MX', 'TXT', 'CAA']:
                continue
            
            if delete_record(record['id'], record['name'], record['type']):
                stats['deleted'] += 1
    
    # Summary
    print("\n📊 Summary:")
    print(f"  Created: {stats['created']}")
    print(f"  Updated: {stats['updated']}")
    print(f"  Deleted: {stats['deleted']}")
    print(f"  Unchanged: {stats['unchanged']}")
    print("\n✅ DNS sync completed!")

def main():
    """Main entry point"""
    if not CF_API_TOKEN or not CF_ZONE_ID:
        print("❌ Error: CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID must be set")
        sys.exit(1)
    
    try:
        sync_dns_records()
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()