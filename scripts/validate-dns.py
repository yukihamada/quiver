#!/usr/bin/env python3
"""
Validate DNS configuration file
"""

import sys
import yaml
import ipaddress
import re
from typing import Dict, List, Any

def validate_record(record: Dict[str, Any], domain: str) -> List[str]:
    """Validate a single DNS record"""
    errors = []
    
    # Required fields
    if 'name' not in record:
        errors.append("Missing 'name' field")
    if 'type' not in record:
        errors.append("Missing 'type' field")
    
    # Validate record type
    valid_types = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'SRV', 'NS']
    if record.get('type') not in valid_types:
        errors.append(f"Invalid record type: {record.get('type')}")
    
    # Type-specific validation
    record_type = record.get('type', '')
    
    if record_type == 'A':
        # Validate IPv4 address
        if 'content' in record:
            try:
                ipaddress.IPv4Address(record['content'])
            except ValueError:
                errors.append(f"Invalid IPv4 address: {record['content']}")
    
    elif record_type == 'AAAA':
        # Validate IPv6 address
        if 'content' in record:
            try:
                ipaddress.IPv6Address(record['content'])
            except ValueError:
                errors.append(f"Invalid IPv6 address: {record['content']}")
    
    elif record_type == 'CNAME':
        # Validate CNAME content
        if 'content' in record:
            if not re.match(r'^[a-zA-Z0-9.-]+$', record['content']):
                errors.append(f"Invalid CNAME target: {record['content']}")
            
            # CNAME at root is not allowed (except with Cloudflare flattening)
            if record.get('name') == '@' and not record.get('proxied', True):
                errors.append("CNAME at root (@) requires Cloudflare proxy")
    
    elif record_type == 'MX':
        # Validate MX record
        if 'priority' not in record:
            errors.append("MX record missing 'priority'")
        elif not isinstance(record['priority'], int) or record['priority'] < 0:
            errors.append(f"Invalid MX priority: {record['priority']}")
    
    elif record_type == 'SRV':
        # Validate SRV record
        required_srv = ['priority', 'weight', 'port', 'target']
        for field in required_srv:
            if field not in record:
                errors.append(f"SRV record missing '{field}'")
        
        # Validate SRV name format
        if 'name' in record:
            if not re.match(r'^_[a-z]+\._[a-z]+$', record['name']):
                errors.append(f"Invalid SRV name format: {record['name']}")
    
    # Validate proxied setting
    if 'proxied' in record:
        if not isinstance(record['proxied'], bool):
            errors.append("'proxied' must be boolean")
        
        # Only certain record types can be proxied
        if record['proxied'] and record_type not in ['A', 'AAAA', 'CNAME']:
            errors.append(f"{record_type} records cannot be proxied")
    
    # Validate TTL
    if 'ttl' in record and record['ttl'] != 'auto':
        if not isinstance(record['ttl'], int) or record['ttl'] < 60:
            errors.append(f"Invalid TTL: {record['ttl']} (must be >= 60 or 'auto')")
    
    return errors

def validate_dns_config(file_path: str = 'dns/records.yaml') -> bool:
    """Validate the entire DNS configuration"""
    print("🔍 Validating DNS configuration...")
    
    try:
        with open(file_path, 'r') as f:
            config = yaml.safe_load(f)
    except Exception as e:
        print(f"❌ Failed to load config file: {e}")
        return False
    
    # Validate required fields
    if 'domain' not in config:
        print("❌ Missing 'domain' field in configuration")
        return False
    
    domain = config['domain']
    if not re.match(r'^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$', domain):
        print(f"❌ Invalid domain: {domain}")
        return False
    
    # Track all errors
    all_errors = []
    record_count = 0
    
    # Validate each section
    for section in ['frontend', 'backend', 'bootstrap', 'srv']:
        if section not in config or not config[section]:
            continue
        
        print(f"\n📁 Validating {section} records...")
        
        for i, record in enumerate(config[section]):
            record_count += 1
            errors = validate_record(record, domain)
            
            if errors:
                all_errors.append({
                    'section': section,
                    'index': i,
                    'record': record,
                    'errors': errors
                })
            else:
                print(f"  ✓ {record.get('name', 'unnamed')} ({record.get('type', 'unknown')})")
    
    # Check for duplicate records
    seen_records = set()
    for section in ['frontend', 'backend', 'bootstrap', 'srv']:
        if section not in config or not config[section]:
            continue
        
        for record in config[section]:
            key = f"{record.get('name')}:{record.get('type')}"
            if key in seen_records:
                all_errors.append({
                    'section': section,
                    'record': record,
                    'errors': [f"Duplicate record: {key}"]
                })
            seen_records.add(key)
    
    # Report results
    print(f"\n📊 Validated {record_count} records")
    
    if all_errors:
        print(f"\n❌ Found {len(all_errors)} invalid records:\n")
        for error in all_errors:
            print(f"Section: {error['section']}")
            if 'index' in error:
                print(f"Index: {error['index']}")
            print(f"Record: {error['record']}")
            for e in error['errors']:
                print(f"  - {e}")
            print()
        return False
    else:
        print("\n✅ All DNS records are valid!")
        return True

def main():
    """Main entry point"""
    if validate_dns_config():
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == '__main__':
    main()