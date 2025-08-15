package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/quiver/provider/internal/models"
)

func getSystemRAM() int {
	// Simplified RAM detection
	if runtime.GOOS == "darwin" {
		// macOS: use sysctl
		// This is simplified - in production, would execute sysctl command
		return 16 // Default assumption
	} else if runtime.GOOS == "linux" {
		// Linux: read from /proc/meminfo
		// This is simplified - in production, would read the file
		return 16 // Default assumption
	}
	return 8
}

func selectModel() {
	fmt.Println("\n🔍 QUIVer モデル選択ツール")
	fmt.Println("==========================")
	
	// Get system RAM
	systemRAM := getSystemRAM()
	fmt.Printf("\n💻 検出されたシステムRAM: %dGB\n", systemRAM)
	
	// Get compatible models
	compatibleModels := models.GetModelsByRAM(systemRAM)
	
	fmt.Println("\n📋 利用可能なモデル:")
	fmt.Println()
	
	// Group by category
	categories := map[string][]models.ModelInfo{
		"general":      {},
		"coding":       {},
		"long-context": {},
	}
	
	for _, model := range compatibleModels {
		categories[model.Category] = append(categories[model.Category], model)
	}
	
	// Display models by category
	modelIndex := 1
	modelMap := make(map[int]models.ModelInfo)
	
	for category, modelList := range categories {
		if len(modelList) == 0 {
			continue
		}
		
		categoryName := map[string]string{
			"general":      "汎用モデル",
			"coding":       "コーディング特化",
			"long-context": "長文対応",
		}[category]
		
		fmt.Printf("【%s】\n", categoryName)
		for _, model := range modelList {
			fmt.Printf("  %d. %s (%dGB RAM必要)\n", modelIndex, model.DisplayName, model.MinRAMGB)
			fmt.Printf("     %s\n", model.Description)
			modelMap[modelIndex] = model
			modelIndex++
		}
		fmt.Println()
	}
	
	// User selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("モデル番号を選択してください (1-" + strconv.Itoa(len(modelMap)) + "): ")
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(modelMap) {
		fmt.Println("❌ 無効な選択です")
		return
	}
	
	selectedModel := modelMap[selection]
	
	fmt.Printf("\n✅ 選択されたモデル: %s\n", selectedModel.DisplayName)
	fmt.Printf("   コンテキストサイズ: %d tokens\n", selectedModel.ContextSize)
	fmt.Printf("   必要RAM: %dGB\n", selectedModel.MinRAMGB)
	
	// Installation command
	fmt.Println("\n🚀 インストールコマンド:")
	fmt.Printf("\n   export QUIVER_MODEL=\"%s\"\n", selectedModel.Name)
	fmt.Println("   curl -fsSL https://quiver.network/install.sh | bash")
	fmt.Println()
	
	// Save selection
	fmt.Print("この設定を保存しますか？ (y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	
	if confirm == "y" || confirm == "yes" {
		// Save to config file
		configPath := os.ExpandEnv("$HOME/.quiver/config/selected_model")
		os.MkdirAll(os.ExpandEnv("$HOME/.quiver/config"), 0755)
		
		err := os.WriteFile(configPath, []byte(selectedModel.Name), 0644)
		if err == nil {
			fmt.Println("✅ 設定が保存されました")
		} else {
			fmt.Printf("⚠️  設定の保存に失敗: %v\n", err)
		}
	}
}

// Add this to main.go as a command option
func init() {
	if len(os.Args) > 1 && os.Args[1] == "select-model" {
		selectModel()
		os.Exit(0)
	}
}