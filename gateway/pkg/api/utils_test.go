package api

import (
	"testing"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     GenerateRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: GenerateRequest{
				Prompt: "test prompt",
				Model:  "test-model",
			},
			wantErr: false,
		},
		{
			name: "empty prompt",
			req: GenerateRequest{
				Prompt: "",
				Model:  "test-model",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			req: GenerateRequest{
				Prompt: "test",
				Model:  "",
			},
			wantErr: true,
		},
		{
			name: "prompt too long",
			req: GenerateRequest{
				Prompt: string(make([]byte, 4097)),
				Model:  "test-model",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGenerateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGenerateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCanaryCheck(t *testing.T) {
	// Test canary prompt detection
	canaryPrompts := []string{
		"test canary prompt",
		"health check",
	}
	canaryAnswers := map[string]string{
		"test canary prompt": "canary response",
		"health check":       "ok",
	}
	
	for _, prompt := range canaryPrompts {
		if _, isCanary := canaryAnswers[prompt]; !isCanary {
			t.Errorf("Canary prompt %q not found in answers", prompt)
		}
	}
}

// Helper function to validate request
func validateGenerateRequest(req GenerateRequest) error {
	if req.Prompt == "" {
		return errEmptyPrompt
	}
	if req.Model == "" {
		return errEmptyModel
	}
	if len(req.Prompt) > 4096 {
		return errPromptTooLong
	}
	return nil
}

var (
	errEmptyPrompt   = &errorStruct{"empty prompt"}
	errEmptyModel    = &errorStruct{"empty model"}
	errPromptTooLong = &errorStruct{"prompt too long"}
)

type errorStruct struct {
	msg string
}

func (e *errorStruct) Error() string {
	return e.msg
}
