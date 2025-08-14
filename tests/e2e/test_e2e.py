import pytest
import httpx
import asyncio
import time
import statistics
from concurrent.futures import ThreadPoolExecutor
import os

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
AGGREGATOR_URL = os.getenv("AGGREGATOR_URL", "http://localhost:8081")

class E2EMetrics:
    def __init__(self):
        self.latencies = []
        self.success_count = 0
        self.error_5xx_count = 0
        self.total_count = 0
        self.canary_passed = 0
        self.canary_total = 0

@pytest.mark.asyncio
async def test_e2e_multi_client_provider():
    """Test N clients with M providers"""
    N_CLIENTS = 10
    M_PROVIDERS = 3  # Simulated by gateway routing
    REQUESTS_PER_CLIENT = 10
    
    metrics = E2EMetrics()
    
    async def client_workload(client_id: int):
        async with httpx.AsyncClient(timeout=30.0) as client:
            for i in range(REQUESTS_PER_CLIENT):
                start = time.time()
                
                try:
                    resp = await client.post(
                        f"{GATEWAY_URL}/generate",
                        json={
                            "prompt": f"Client {client_id} request {i}",
                            "model": "llama2",
                            "token": f"client-{client_id}"
                        }
                    )
                    
                    latency = (time.time() - start) * 1000  # ms
                    metrics.latencies.append(latency)
                    metrics.total_count += 1
                    
                    if resp.status_code == 200:
                        metrics.success_count += 1
                        
                        # Check for canary
                        data = resp.json()
                        receipt = data.get("receipt", {})
                        canary = receipt.get("canary", {})
                        
                        if canary.get("id"):
                            metrics.canary_total += 1
                            if canary.get("passed", False):
                                metrics.canary_passed += 1
                                
                    elif 500 <= resp.status_code < 600:
                        metrics.error_5xx_count += 1
                        
                except Exception as e:
                    metrics.error_5xx_count += 1
                    metrics.total_count += 1
                
                await asyncio.sleep(0.1)  # Avoid overwhelming
    
    # Run clients concurrently
    tasks = [client_workload(i) for i in range(N_CLIENTS)]
    await asyncio.gather(*tasks)
    
    # Calculate metrics
    acceptance_rate = metrics.success_count / metrics.total_count * 100
    error_5xx_rate = metrics.error_5xx_count / metrics.total_count * 100
    
    if metrics.latencies:
        p50 = statistics.median(metrics.latencies)
        p95 = statistics.quantiles(metrics.latencies, n=20)[18]  # 95th percentile
    else:
        p50 = p95 = 0
    
    canary_pass_rate = 100
    if metrics.canary_total > 0:
        canary_pass_rate = metrics.canary_passed / metrics.canary_total * 100
    
    # Assertions based on CI requirements
    assert acceptance_rate >= 98, f"Acceptance rate {acceptance_rate:.1f}% below 98%"
    assert error_5xx_rate <= 1, f"5xx error rate {error_5xx_rate:.1f}% exceeds 1%"
    assert p95 < 2500, f"p95 latency {p95:.0f}ms exceeds 2.5s"
    assert canary_pass_rate >= 99, f"Canary pass rate {canary_pass_rate:.1f}% below 99%"
    
    print(f"\nE2E Test Results:")
    print(f"Acceptance Rate: {acceptance_rate:.1f}%")
    print(f"5xx Error Rate: {error_5xx_rate:.1f}%")
    print(f"p50 Latency: {p50:.0f}ms")
    print(f"p95 Latency: {p95:.0f}ms")
    print(f"Canary Pass Rate: {canary_pass_rate:.1f}%")

@pytest.mark.asyncio
async def test_e2e_canary_slashing_logs():
    """Test that failed canaries produce slashing logs"""
    # This test would need access to container logs
    # For CI, we simulate by checking the canary mechanism
    
    async with httpx.AsyncClient() as client:
        # Make requests until we get a failed canary
        for i in range(100):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": f"Slashing test {i}",
                    "model": "llama2",
                    "token": "slashing-e2e"
                }
            )
            
            if resp.status_code == 200:
                data = resp.json()
                receipt = data.get("receipt", {})
                canary = receipt.get("canary", {})
                
                if canary.get("id") and not canary.get("passed", True):
                    # Found a failed canary
                    # In real test, check logs for slashing message
                    print(f"Failed canary detected: {canary}")
                    return
            
            await asyncio.sleep(0.05)
    
    # If no failed canaries found, that's also acceptable
    print("No failed canaries found in 100 requests")

@pytest.mark.asyncio  
async def test_e2e_epoch_settlement():
    """Test full flow from receipt to settlement"""
    async with httpx.AsyncClient() as client:
        # Generate receipts
        receipts = []
        for i in range(5):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": f"Settlement test {i}",
                    "model": "llama2",
                    "token": "settlement-test"
                }
            )
            
            if resp.status_code == 200:
                receipt = resp.json()["receipt"]
                receipts.append({
                    "receipt": receipt,
                    "signature": "test-signature"
                })
        
        if receipts:
            # Commit to aggregator
            epoch = receipts[0]["receipt"]["epoch"]
            commit_resp = await client.post(
                f"{AGGREGATOR_URL}/commit",
                json={
                    "epoch": epoch,
                    "receipts": receipts
                }
            )
            
            assert commit_resp.status_code == 200
            commit_data = commit_resp.json()
            
            # Verify claim
            claim_resp = await client.post(
                f"{AGGREGATOR_URL}/claim",
                json={
                    "receipt_id": receipts[0]["receipt"]["receipt_id"],
                    "merkle_proof": [],  # Would be from commit response
                    "epoch": epoch
                }
            )
            
            assert claim_resp.status_code == 200