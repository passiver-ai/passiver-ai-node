from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from vllm import LLM, SamplingParams

app = FastAPI()
llm = LLM(model="facebook/opt-125m")

class GenerateRequest(BaseModel):
    prompt: str
    max_tokens: int = 64
    temperature: float = 0.8
    top_p: float = 0.95

@app.post("/generate")
async def generate(request: GenerateRequest):
    try:
        sampling_params = SamplingParams(
            temperature=request.temperature,
            top_p=request.top_p,
            max_tokens=request.max_tokens
        )
        outputs = llm.generate([request.prompt], sampling_params)
        generated_text = outputs[0].outputs[0].text
        return {"generated_text": generated_text}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def root():
    return {"message": "vLLM API is running"}

@app.get("/health")
async def health():
    return {"status": "healthy"}
