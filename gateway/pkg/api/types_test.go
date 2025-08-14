package api

import (
	"encoding/json"
	"testing"
)

func TestGenerateRequestJSON(t *testing.T) {
	req := GenerateRequest{
		Prompt: "test prompt",
		Model:  "test-model",
		Token:  "test-token",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	var decoded GenerateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Prompt != req.Prompt {
		t.Errorf("Expected prompt %q, got %q", req.Prompt, decoded.Prompt)
	}

	if decoded.Model != req.Model {
		t.Errorf("Expected model %q, got %q", req.Model, decoded.Model)
	}

	if decoded.Token != req.Token {
		t.Errorf("Expected token %q, got %q", req.Token, decoded.Token)
	}
}

func TestGenerateResponseJSON(t *testing.T) {
	resp := GenerateResponse{
		Completion: "test completion",
		Receipt: map[string]interface{}{
			"id": "test-id",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded GenerateResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Completion != resp.Completion {
		t.Errorf("Expected completion %q, got %q", resp.Completion, decoded.Completion)
	}

	if decoded.Receipt == nil {
		t.Error("Expected receipt")
	}
}

func TestErrorResponseJSON(t *testing.T) {
	resp := ErrorResponse{
		Error: "test error",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Error != resp.Error {
		t.Errorf("Expected error %q, got %q", resp.Error, decoded.Error)
	}
}
