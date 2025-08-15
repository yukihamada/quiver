package models

import (
	"fmt"
	"strings"
)

// ModelInfo represents information about a supported model
type ModelInfo struct {
	Name         string
	DisplayName  string
	MinRAMGB     int
	Description  string
	ContextSize  int
	Category     string
}

// SupportedModels is the registry of all supported models
var SupportedModels = map[string]ModelInfo{
	// Qwen3 Series
	"qwen3:0.6b": {
		Name:         "qwen3:0.6b",
		DisplayName:  "Qwen3 0.6B",
		MinRAMGB:     2,
		Description:  "Ultra-lightweight model for basic tasks",
		ContextSize:  8192,
		Category:     "general",
	},
	"qwen3:3b": {
		Name:         "qwen3:3b",
		DisplayName:  "Qwen3 3B",
		MinRAMGB:     8,
		Description:  "Balanced performance and efficiency",
		ContextSize:  8192,
		Category:     "general",
	},
	"qwen3:7b": {
		Name:         "qwen3:7b",
		DisplayName:  "Qwen3 7B",
		MinRAMGB:     16,
		Description:  "High performance for complex tasks",
		ContextSize:  8192,
		Category:     "general",
	},
	"qwen3:14b": {
		Name:         "qwen3:14b",
		DisplayName:  "Qwen3 14B",
		MinRAMGB:     32,
		Description:  "Professional grade model",
		ContextSize:  8192,
		Category:     "general",
	},
	"qwen3:32b": {
		Name:         "qwen3:32b",
		DisplayName:  "Qwen3 32B",
		MinRAMGB:     64,
		Description:  "Enterprise level performance",
		ContextSize:  8192,
		Category:     "general",
	},
	"qwen3-coder:30b": {
		Name:         "qwen3-coder:30b",
		DisplayName:  "Qwen3 Coder 30B",
		MinRAMGB:     64,
		Description:  "Specialized for code generation",
		ContextSize:  16384,
		Category:     "coding",
	},
	
	// GPT-OSS Series
	"gpt-oss:20b": {
		Name:         "gpt-oss:20b",
		DisplayName:  "GPT-OSS 20B",
		MinRAMGB:     48,
		Description:  "Open source GPT for local execution",
		ContextSize:  8192,
		Category:     "general",
	},
	"gpt-oss:120b": {
		Name:         "gpt-oss:120b",
		DisplayName:  "GPT-OSS 120B",
		MinRAMGB:     256,
		Description:  "Flagship open source model",
		ContextSize:  8192,
		Category:     "general",
	},
	
	// Jan-Nano Series
	"jan-nano:32k": {
		Name:         "jan-nano:32k",
		DisplayName:  "Jan-Nano 32K",
		MinRAMGB:     16,
		Description:  "Long context support (32K tokens)",
		ContextSize:  32768,
		Category:     "long-context",
	},
	"jan-nano:128k": {
		Name:         "jan-nano:128k",
		DisplayName:  "Jan-Nano 128K",
		MinRAMGB:     32,
		Description:  "Ultra-long context support (128K tokens)",
		ContextSize:  131072,
		Category:     "long-context",
	},
	
	// Other popular models
	"llama3.2:3b": {
		Name:         "llama3.2:3b",
		DisplayName:  "Llama 3.2 3B",
		MinRAMGB:     8,
		Description:  "Meta's efficient model",
		ContextSize:  8192,
		Category:     "general",
	},
	"mistral:7b": {
		Name:         "mistral:7b",
		DisplayName:  "Mistral 7B",
		MinRAMGB:     16,
		Description:  "Efficient European model",
		ContextSize:  8192,
		Category:     "general",
	},
	"phi3:mini": {
		Name:         "phi3:mini",
		DisplayName:  "Phi-3 Mini",
		MinRAMGB:     4,
		Description:  "Microsoft's compact model",
		ContextSize:  4096,
		Category:     "general",
	},
	"gemma2:9b": {
		Name:         "gemma2:9b",
		DisplayName:  "Gemma 2 9B",
		MinRAMGB:     20,
		Description:  "Google's efficient model",
		ContextSize:  8192,
		Category:     "general",
	},
}

// GetModelInfo returns information about a specific model
func GetModelInfo(modelName string) (ModelInfo, error) {
	// Normalize model name
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	
	if info, exists := SupportedModels[modelName]; exists {
		return info, nil
	}
	
	return ModelInfo{}, fmt.Errorf("model %s not found in registry", modelName)
}

// GetModelsByCategory returns all models in a specific category
func GetModelsByCategory(category string) []ModelInfo {
	var models []ModelInfo
	for _, model := range SupportedModels {
		if model.Category == category {
			models = append(models, model)
		}
	}
	return models
}

// GetModelsByRAM returns models that can run with the given RAM
func GetModelsByRAM(availableRAMGB int) []ModelInfo {
	var models []ModelInfo
	for _, model := range SupportedModels {
		if model.MinRAMGB <= availableRAMGB {
			models = append(models, model)
		}
	}
	return models
}

// ValidateModel checks if a model is supported and system meets requirements
func ValidateModel(modelName string, systemRAMGB int) error {
	info, err := GetModelInfo(modelName)
	if err != nil {
		return err
	}
	
	if systemRAMGB < info.MinRAMGB {
		return fmt.Errorf("insufficient RAM for %s: need %dGB, have %dGB", 
			info.DisplayName, info.MinRAMGB, systemRAMGB)
	}
	
	return nil
}