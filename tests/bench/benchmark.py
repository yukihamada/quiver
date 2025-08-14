import asyncio
import httpx
import time
import statistics
import json
from datetime import datetime
import sys

GATEWAY_URL = "http://localhost:8080"
WORKERS = 2
MAX_TOKENS = 256

class BenchmarkResults:
    def __init__(self):
        self.latencies = []
        self.errors = []
        self.start_time = None
        self.end_time = None

async def worker(worker_id: int, num_requests: int, results: BenchmarkResults):
    async with httpx.AsyncClient(timeout=30.0) as client:
        for i in range(num_requests):
            prompt = f"Benchmark test {worker_id}-{i}"
            
            start = time.time()
            try:
                resp = await client.post(
                    f"{GATEWAY_URL}/generate",
                    json={
                        "prompt": prompt[:256],  # Keep under token limit
                        "model": "llama2",
                        "token": f"bench-{worker_id}"
                    }
                )
                
                latency = (time.time() - start) * 1000  # ms
                results.latencies.append(latency)
                
                if resp.status_code >= 500:
                    results.errors.append(resp.status_code)
                    
            except Exception as e:
                results.errors.append(str(e))
                results.latencies.append(30000)  # timeout
            
            await asyncio.sleep(0.01)  # Small delay between requests

async def run_benchmark(total_requests: int = 100):
    results = BenchmarkResults()
    results.start_time = datetime.now()
    
    requests_per_worker = total_requests // WORKERS
    
    # Run workers concurrently
    tasks = [
        worker(i, requests_per_worker, results) 
        for i in range(WORKERS)
    ]
    
    await asyncio.gather(*tasks)
    
    results.end_time = datetime.now()
    
    # Calculate metrics
    if results.latencies:
        p50 = statistics.median(results.latencies)
        p95 = statistics.quantiles(results.latencies, n=20)[18]
        p99 = statistics.quantiles(results.latencies, n=100)[98]
        avg = statistics.mean(results.latencies)
        
        duration = (results.end_time - results.start_time).total_seconds()
        throughput = len(results.latencies) / duration
        
        error_rate = len(results.errors) / (len(results.latencies) + len(results.errors)) * 100
        
        report = {
            "timestamp": results.start_time.isoformat(),
            "duration_seconds": duration,
            "total_requests": len(results.latencies) + len(results.errors),
            "successful_requests": len(results.latencies),
            "errors": len(results.errors),
            "error_rate_percent": error_rate,
            "throughput_rps": throughput,
            "latency_ms": {
                "avg": avg,
                "p50": p50,
                "p95": p95,
                "p99": p99,
                "min": min(results.latencies),
                "max": max(results.latencies)
            },
            "workers": WORKERS,
            "ci_pass": {
                "p95_under_2500ms": p95 < 2500,
                "error_rate_under_1pct": error_rate <= 1
            }
        }
        
        print(json.dumps(report, indent=2))
        
        # Write to file for CI
        with open("benchmark_results.json", "w") as f:
            json.dump(report, f, indent=2)
        
        # Exit code for CI
        if not report["ci_pass"]["p95_under_2500ms"]:
            print(f"\nFAILED: p95 latency {p95:.0f}ms exceeds 2500ms", file=sys.stderr)
            sys.exit(1)
        if not report["ci_pass"]["error_rate_under_1pct"]:
            print(f"\nFAILED: Error rate {error_rate:.1f}% exceeds 1%", file=sys.stderr)
            sys.exit(1)
            
        print("\nBenchmark PASSED all CI criteria")
        
    else:
        print("No successful requests completed", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    requests = int(sys.argv[1]) if len(sys.argv) > 1 else 100
    asyncio.run(run_benchmark(requests))