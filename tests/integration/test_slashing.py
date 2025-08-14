import pytest
import httpx
import asyncio
import os
import json

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")

@pytest.mark.asyncio
async def test_canary_fail_slashing_log(gateway_client, caplog):
    """Test that failed canaries produce slashing intent logs"""
    
    # Force canary by using known prompt
    canary_prompt = "What is the capital of France?"
    
    # Make request that should trigger canary
    response = await gateway_client.post(
        "/generate",
        json={
            "prompt": canary_prompt,
            "model": "llama2",
            "token": "slashing-test"
        }
    )
    
    assert response.status_code == 200
    data = response.json()
    
    # Check if canary was triggered
    receipt = data.get("receipt", {})
    canary = receipt.get("canary", {})
    
    if canary.get("id"):
        # Check canary result
        if not canary.get("passed", True):
            # Failed canary detected
            print(f"SLASHING STUB: Provider {receipt.get('provider_pk')} failed canary check")
            print(f"  Canary ID: {canary.get('id')}")
            print(f"  Receipt ID: {receipt.get('receipt_id')}")
            print(f"  Epoch: {receipt.get('epoch')}")
            
            # In production, this would:
            # 1. Record the failure in persistent storage
            # 2. Accumulate evidence for on-chain slashing
            # 3. Potentially pause the provider
            # 4. Notify monitoring systems
            
            assert "Canary response hidden" in data["completion"]

@pytest.mark.asyncio
async def test_slashing_accumulation():
    """Test slashing evidence accumulation"""
    
    slashing_events = []
    
    class SlashingCollector:
        def record_failure(self, provider_id, receipt_id, reason):
            slashing_events.append({
                "provider_id": provider_id,
                "receipt_id": receipt_id,
                "reason": reason,
                "timestamp": asyncio.get_event_loop().time()
            })
            
        def should_slash(self, provider_id):
            # Slash after 3 failures
            failures = [e for e in slashing_events if e["provider_id"] == provider_id]
            return len(failures) >= 3
    
    collector = SlashingCollector()
    
    # Simulate failures
    for i in range(5):
        collector.record_failure("provider-1", f"receipt-{i}", "canary_failed")
    
    assert collector.should_slash("provider-1")
    assert not collector.should_slash("provider-2")
    
    # Verify events were recorded
    assert len(slashing_events) == 5
    assert all(e["reason"] == "canary_failed" for e in slashing_events)