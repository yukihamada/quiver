import pytest
import httpx
import asyncio
import os
from collections import Counter

GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")

@pytest.mark.asyncio
async def test_redundant_execution_2of3():
    """Test 2-of-3 redundant execution sampling"""
    async with httpx.AsyncClient() as client:
        mismatches = 0
        total_checks = 50
        
        for i in range(total_checks):
            prompt = f"Redundancy test prompt {i}"
            
            # Simulate 3 provider responses
            responses = []
            for j in range(3):
                resp = await client.post(
                    f"{GATEWAY_URL}/generate",
                    json={
                        "prompt": prompt,
                        "model": "llama2",
                        "token": f"redundancy-{j}"
                    }
                )
                
                if resp.status_code == 200:
                    responses.append(resp.json()["completion"])
                
                await asyncio.sleep(0.05)  # Avoid rate limit
            
            if len(responses) == 3:
                # Check for consensus (2-of-3)
                counter = Counter(responses)
                most_common = counter.most_common(1)[0]
                
                if most_common[1] < 2:
                    # No consensus - all different
                    mismatches += 1
                elif most_common[1] == 2:
                    # One outlier
                    # In deterministic mode, this shouldn't happen
                    mismatches += 1
        
        # False positive rate should be very low
        false_positive_rate = mismatches / total_checks
        assert false_positive_rate <= 0.01, f"False positive rate {false_positive_rate} exceeds 1%"

@pytest.mark.asyncio
async def test_redundancy_hash_verification():
    """Test that receipt hashes match for identical inputs"""
    async with httpx.AsyncClient() as client:
        prompt = "Test prompt for hash verification"
        receipts = []
        
        # Get 3 responses for same prompt
        for i in range(3):
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": prompt,
                    "model": "llama2", 
                    "token": "hash-verify"
                }
            )
            
            if resp.status_code == 200:
                receipt = resp.json()["receipt"]
                receipts.append(receipt)
            
            await asyncio.sleep(0.1)
        
        if len(receipts) >= 2:
            # All prompt hashes should be identical
            prompt_hashes = [r["prompt_hash"] for r in receipts]
            assert len(set(prompt_hashes)) == 1, "Prompt hashes differ for same input"
            
            # In deterministic mode, output hashes should also match
            output_hashes = [r["output_hash"] for r in receipts]
            assert len(set(output_hashes)) == 1, "Output hashes differ in deterministic mode"

@pytest.mark.asyncio
async def test_byzantine_fault_detection():
    """Test detection of inconsistent provider responses"""
    async with httpx.AsyncClient() as client:
        # Simulate byzantine behavior by checking response patterns
        responses_by_prompt = {}
        
        for i in range(30):
            prompt = f"Byzantine test {i % 10}"  # Reuse some prompts
            
            resp = await client.post(
                f"{GATEWAY_URL}/generate",
                json={
                    "prompt": prompt,
                    "model": "llama2",
                    "token": "byzantine-test"
                }
            )
            
            if resp.status_code == 200:
                completion = resp.json()["completion"]
                
                if prompt not in responses_by_prompt:
                    responses_by_prompt[prompt] = []
                responses_by_prompt[prompt].append(completion)
            
            await asyncio.sleep(0.05)
        
        # Check consistency for repeated prompts
        inconsistencies = 0
        for prompt, completions in responses_by_prompt.items():
            if len(completions) > 1:
                unique_completions = set(completions)
                if len(unique_completions) > 1:
                    inconsistencies += 1
        
        # In deterministic mode, there should be no inconsistencies
        assert inconsistencies == 0, f"Found {inconsistencies} inconsistent responses"