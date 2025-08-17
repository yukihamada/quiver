#!/usr/bin/env python3
"""
Verify DNS propagation after sync
"""

import dns.resolver
import yaml
import time
import sys
from typing import Dict, List, Any

def load_dns_config(file_path: str = 'dns/records.yaml') -> Dict[str, Any]:
    """Load DNS configuration from YAML file"""
    with open(file_path, 'r') as f:
        return yaml.safe_load(f)

def check_dns_record(name: str, record_type: str, expected_content: str) -> bool:
    """Check if a DNS record has propagated correctly"""
    try:
        resolver = dns.resolver.Resolver()
        # Use multiple DNS servers for verification
        resolver.nameservers = ['1.1.1.1', '8.8.8.8', '9.9.9.9']
        
        answers = resolver.resolve(name, record_type)
        
        for rdata in answers:
            if record_type == 'A':
                if str(rdata) == expected_content:
                    return True
            elif record_type == 'CNAME':
                # CNAME records need to match the target (with or without trailing dot)
                if str(rdata).rstrip('.') == expected_content.rstrip('.'):
                    return True
            elif record_type == 'TXT':
                # TXT records might have quotes
                if str(rdata).strip('"') == expected_content.strip('"'):
                    return True
        
        return False
    except dns.resolver.NXDOMAIN:
        print(f"  ⚠️  {name} does not exist")
        return False
    except dns.resolver.NoAnswer:
        print(f"  ⚠️  No {record_type} record found for {name}")
        return False
    except Exception as e:
        print(f"  ❌ Error checking {name}: {e}")
        return False

def verify_dns_propagation():
    """Verify DNS propagation for all records"""
    print("🔍 Verifying DNS propagation...")
    print("   Using DNS servers: 1.1.1.1, 8.8.8.8, 9.9.9.9\n")
    
    # Load configuration
    config = load_dns_config()
    domain = config['domain']
    
    # Collect all records to verify
    records_to_check = []
    
    for section in ['frontend', 'backend', 'bootstrap', 'srv']:
        if section not in config or not config[section]:
            continue
        
        for record in config[section]:
            # Skip SRV records for now (different format)
            if record['type'] == 'SRV':
                continue
            
            # Handle @ for root domain
            if record['name'] == '@':
                full_name = domain
            else:
                full_name = f"{record['name']}.{domain}"
            
            records_to_check.append({
                'name': full_name,
                'type': record['type'],
                'content': record['content']
            })
    
    # Check each record
    total = len(records_to_check)
    success = 0
    failed = []
    
    print(f"Checking {total} DNS records...\n")
    
    for i, record in enumerate(records_to_check, 1):
        print(f"[{i}/{total}] Checking {record['name']} ({record['type']})...", end=' ')
        
        if check_dns_record(record['name'], record['type'], record['content']):
            print("✅")
            success += 1
        else:
            print(f"❌ Expected: {record['content']}")
            failed.append(record)
        
        # Small delay to avoid rate limiting
        if i < total:
            time.sleep(0.5)
    
    # Summary
    print(f"\n📊 Summary:")
    print(f"  Total records: {total}")
    print(f"  Successful: {success}")
    print(f"  Failed: {len(failed)}")
    
    if failed:
        print(f"\n❌ The following records have not propagated yet:")
        for record in failed:
            print(f"  - {record['name']} ({record['type']}) → {record['content']}")
        print(f"\n⏳ DNS propagation can take up to 48 hours. Try again later.")
        return False
    else:
        print(f"\n✅ All DNS records have propagated successfully!")
        return True

def main():
    """Main entry point"""
    try:
        if verify_dns_propagation():
            sys.exit(0)
        else:
            sys.exit(1)
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()