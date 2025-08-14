import pytest
import httpx
import asyncio
import json
import os
from typing import List, Dict

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
AGGREGATOR_URL = os.getenv("AGGREGATOR_URL", "http://localhost:8081")

@pytest.mark.asyncio
async def test_provider_gateway_aggregator_flow():
    async with httpx.AsyncClient() as client:
        # Step 1: Generate completion via gateway
        response = await client.post(
            f"{GATEWAY_URL}/generate",
            json={
                "prompt": "Test prompt for integration",
                "model": "llama2",
                "token": "test-token-integration"
            }
        )
        assert response.status_code == 200
        data = response.json()
        assert "completion" in data
        assert "receipt" in data
        
        receipt = data["receipt"]
        assert receipt["version"] == "1.0.0"
        assert receipt["tokens_in"] > 0
        assert receipt["tokens_out"] > 0
        
        # Step 2: Collect multiple receipts
        receipts = [receipt]
        for i in range(4):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": f"Test prompt {i}",
                    "model": "llama2",
                    "token": "test-token-integration"
                }
            )
            if resp.status_code == 200:
                receipts.append(resp.json()["receipt"])
        
        # Step 3: Commit to aggregator
        epoch = receipts[0]["epoch"]
        commit_resp = await client.post(
            f"{AGGREGATOR_URL}/commit",
            json={
                "epoch": epoch,
                "receipts": [{"receipt": r, "signature": "test-sig"} for r in receipts]
            }
        )
        assert commit_resp.status_code == 200
        commit_data = commit_resp.json()
        assert "merkle_root" in commit_data
        assert commit_data["receipt_count"] == len(receipts)

@pytest.mark.asyncio
async def test_deterministic_responses():
    async with httpx.AsyncClient() as client:
        prompt = "What is the meaning of life?"
        responses = []
        
        # Make same request 5 times
        for _ in range(5):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": prompt,
                    "model": "llama2",
                    "token": "test-determinism"
                }
            )
            if resp.status_code == 200:
                responses.append(resp.json()["completion"])
        
        # All responses should be identical
        assert len(set(responses)) == 1, "Responses not deterministic"

@pytest.mark.asyncio
async def test_rate_limiting():
    async with httpx.AsyncClient() as client:
        token = "rate-limit-test"
        successful = 0
        rate_limited = 0
        
        # Burst 30 requests
        tasks = []
        for i in range(30):
            task = client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": f"Rate limit test {i}",
                    "model": "llama2",
                    "token": token
                }
            )
            tasks.append(task)
        
        responses = await asyncio.gather(*tasks, return_exceptions=True)
        
        for resp in responses:
            if isinstance(resp, httpx.Response):
                if resp.status_code == 200:
                    successful += 1
                elif resp.status_code == 429:
                    rate_limited += 1
        
        assert successful > 0, "No requests succeeded"
        assert rate_limited > 0, "No rate limiting occurred"

@pytest.mark.asyncio
async def test_prompt_size_limit():
    async with httpx.AsyncClient() as client:
        # Test exactly at limit
        prompt_4096 = "a" * 4096
        resp = await client.post(
            f"{GATEWAY_URL}/generate",
            json={
                "prompt": prompt_4096,
                "model": "llama2",
                "token": "size-test"
            }
        )
        assert resp.status_code == 200
        
        # Test over limit
        prompt_4097 = "a" * 4097
        resp = await client.post(
            f"{GATEWAY_URL}/generate",
            json={
                "prompt": prompt_4097,
                "model": "llama2",
                "token": "size-test"
            }
        )
        assert resp.status_code == 400