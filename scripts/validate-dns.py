#!/usr/bin/env python3
"""Validate DNS records configuration."""

import yaml
import sys
from pathlib import Path

def validate_dns_config():
    """Validate the DNS records YAML configuration."""
    config_path = Path('dns/records.yaml')
    
    if not config_path.exists():
        print(f"❌ DNS config file not found: {config_path}")
        return False
    
    try:
        with open(config_path, 'r') as f:
            config = yaml.safe_load(f)
    except Exception as e:
        print(f"❌ Failed to parse YAML: {e}")
        return False
    
    # Validate required fields
    if 'domain' not in config:
        print("❌ Missing 'domain' field in config")
        return False
    
    valid = True
    record_count = 0
    
    # Validate each section
    for section in ['frontend', 'backend']:
        if section not in config:
            continue
            
        for record in config[section]:
            record_count += 1
            
            # Check required fields
            for field in ['name', 'type', 'content']:
                if field not in record:
                    print(f"❌ Missing '{field}' in record: {record}")
                    valid = False
            
            # Validate record type
            valid_types = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SRV']
            if record.get('type') not in valid_types:
                print(f"❌ Invalid record type '{record.get('type')}' in: {record}")
                valid = False
            
            # Validate CNAME records don't conflict with other records at root
            if record.get('type') == 'CNAME' and record.get('name') == '@':
                print(f"⚠️  Warning: CNAME at root domain (@) may conflict with other records")
    
    print(f"✅ Validated {record_count} DNS records")
    
    if valid:
        print("✅ DNS configuration is valid")
    else:
        print("❌ DNS configuration has errors")
    
    return valid

if __name__ == '__main__':
    sys.exit(0 if validate_dns_config() else 1)