from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import hashlib
import time
import os

app = FastAPI()

class GenerateRequest(BaseModel):
    model: str
    prompt: str
    temperature: float = 0
    seed: int = 42
    stream: bool = False

class GenerateResponse(BaseModel):
    model: str
    response: str
    total_duration: int
    prompt_eval_count: int
    eval_count: int

CANARY_RESPONSES = {
    "What is the capital of France?": "The capital of France is Paris.",
    "Calculate 2 + 2": "2 + 2 equals 4.",
    "Who wrote Romeo and Juliet?": "William Shakespeare wrote Romeo and Juliet."
}

DETERMINISTIC_RESPONSES = {
    "default": "This is a deterministic response for testing."
}

@app.post("/api/generate")
async def generate(request: GenerateRequest):
    if request.temperature != 0:
        raise HTTPException(status_code=400, detail="Temperature must be 0")
    
    if request.seed != 42:
        raise HTTPException(status_code=400, detail="Seed must be 42")
    
    # Check for canary
    if request.prompt in CANARY_RESPONSES:
        response_text = CANARY_RESPONSES[request.prompt]
    else:
        # Generate deterministic response based on prompt hash
        prompt_hash = hashlib.sha256(request.prompt.encode()).hexdigest()
        response_text = f"Deterministic response for hash {prompt_hash[:8]}"
    
    # Simulate processing time
    time.sleep(0.05)
    
    # Calculate token counts deterministically
    prompt_tokens = len(request.prompt.split())
    response_tokens = len(response_text.split())
    
    return GenerateResponse(
        model=request.model,
        response=response_text,
        total_duration=50000000,  # 50ms in nanoseconds
        prompt_eval_count=prompt_tokens,
        eval_count=response_tokens
    )

@app.get("/api/tags")
async def tags():
    return {
        "models": [
            {"name": "llama2", "size": 3825819519},
            {"name": "jan-nano", "size": 1000000000}
        ]
    }

@app.get("/health")
async def health():
    return {"status": "healthy"}