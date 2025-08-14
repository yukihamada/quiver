import pytest
import asyncio
import httpx
import time
import os

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
AGGREGATOR_URL = os.getenv("AGGREGATOR_URL", "http://localhost:8081")

@pytest.fixture
async def gateway_client():
    async with httpx.AsyncClient(base_url=GATEWAY_URL, timeout=30.0) as client:
        # Wait for service to be ready
        for _ in range(10):
            try:
                resp = await client.get("/health")
                if resp.status_code == 200:
                    break
            except:
                pass
            await asyncio.sleep(1)
        yield client

@pytest.fixture
async def aggregator_client():
    async with httpx.AsyncClient(base_url=AGGREGATOR_URL, timeout=30.0) as client:
        # Wait for service to be ready
        for _ in range(10):
            try:
                resp = await client.get("/health")
                if resp.status_code == 200:
                    break
            except:
                pass
            await asyncio.sleep(1)
        yield client

@pytest.fixture
def canary_prompts():
    return [
        "What is the capital of France?",
        "Calculate 2 + 2",
        "Who wrote Romeo and Juliet?"
    ]

@pytest.fixture
def expected_canary_answers():
    return {
        "What is the capital of France?": "Paris",
        "Calculate 2 + 2": "4",
        "Who wrote Romeo and Juliet?": "William Shakespeare"
    }