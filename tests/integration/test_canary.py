import pytest
import httpx
import asyncio
import os
import json

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")

CANARY_PROMPTS = [
    "What is the capital of France?",
    "Calculate 2 + 2",
    "Who wrote Romeo and Juliet?"
]

EXPECTED_ANSWERS = {
    "What is the capital of France?": "Paris",
    "Calculate 2 + 2": "4",
    "Who wrote Romeo and Juliet?": "William Shakespeare"
}

@pytest.mark.asyncio
async def test_canary_execution():
    async with httpx.AsyncClient() as client:
        canary_count = 0
        total_requests = 100
        
        # Make many requests to trigger canaries
        for i in range(total_requests):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": f"Regular prompt {i}",
                    "model": "llama2",
                    "token": "canary-test"
                }
            )
            
            if resp.status_code == 200:
                data = resp.json()
                receipt = data.get("receipt", {})
                canary = receipt.get("canary", {})
                
                if canary.get("id"):
                    canary_count += 1
                    
                    # Verify canary was properly hidden
                    assert data["completion"] == "Canary response hidden"
        
        # Should have ~5% canaries
        canary_rate = canary_count / total_requests
        assert 0.02 <= canary_rate <= 0.08, f"Canary rate {canary_rate} outside expected range"

@pytest.mark.asyncio
async def test_canary_validation():
    async with httpx.AsyncClient(timeout=30.0) as client:
        passed_count = 0
        failed_count = 0
        
        # Force canaries by using known prompts
        for _ in range(20):
            for prompt in CANARY_PROMPTS:
                # Make request that should trigger canary
                resp = await client.post(
                    f"{GATEWAY_URL}/generate",
                    json={
                        "prompt": prompt,
                        "model": "llama2",
                        "token": "canary-validate"
                    }
                )
                
                if resp.status_code == 200:
                    data = resp.json()
                    receipt = data.get("receipt", {})
                    canary = receipt.get("canary", {})
                    
                    if canary.get("passed") is not None:
                        if canary["passed"]:
                            passed_count += 1
                        else:
                            failed_count += 1
                
                await asyncio.sleep(0.1)  # Rate limit
        
        # Canary pass rate should be very high
        if passed_count + failed_count > 0:
            pass_rate = passed_count / (passed_count + failed_count)
            assert pass_rate >= 0.99, f"Canary pass rate {pass_rate} below 99%"

@pytest.mark.asyncio
async def test_canary_slashing_stub():
    """Test that failed canaries log slashing intent"""
    async with httpx.AsyncClient() as client:
        # This would need access to provider logs in real test
        # For now, just verify the canary mechanism works
        
        resp = await client.post(
            f"{GATEWAY_URL}/generate",
            json={
                "prompt": "Test for slashing",
                "model": "llama2",
                "token": "slashing-test"
            }
        )
        
        assert resp.status_code == 200
        # In real implementation, would check logs for slashing messages