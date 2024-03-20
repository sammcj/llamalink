package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ollamaModelsDir = filepath.Join(os.Getenv("HOME"), ".ollama", "models")
var lmStudioModelsDir = filepath.Join(os.Getenv("HOME"), ".cache", "lm-studio", "models")

// print the configured model paths
func printModelPaths() {
	fmt.Println("Ollama models directory:", ollamaModelsDir)
	fmt.Println("LM Studio models directory:", lmStudioModelsDir)
	fmt.Println()
}

func getModelList() ([]string, error) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var models []string
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			models = append(models, fields[0])
		}
	}
	return models, nil
}

func getModelPath(modelName string) (string, error) {
	cmd := exec.Command("ollama", "show", "--modelfile", modelName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "FROM ") {
			return strings.TrimSpace(line[5:]), nil
		}
	}
	return "", fmt.Errorf("model path not found for %s", modelName)
}

func cleanBrokenSymlinks() {
	err := filepath.Walk(lmStudioModelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			_, err := os.Stat(path)
			if err != nil {
				fmt.Printf("Removing broken symlink: %s\n", path)
				os.Remove(path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking LM Studio models directory: %v\n", err)
	}
}

func main() {
	printModelPaths()

	models, err := getModelList()
	if err != nil {
		fmt.Printf("Error getting model list: %v\n", err)
		return
	}

	if len(models) == 0 {
		fmt.Println("No Ollama models found.")
		return
	}

	fmt.Println("\033[1;56mSelect the models to link to LM Studio:\033[0m")
	//"Select the models to link to LM Studio:")
	for i, modelName := range models {
		// colourise the model name
		modelName = "\033[1;36m" + modelName + "\033[0m"

		// colourise the number to red
		fmt.Printf("\033[1;31m%d.\033[0m %s\n", i+1, modelName)
	}

	var selectedModels []int

	fmt.Println()
	fmt.Print("Enter the model numbers (comma-separated), or press Enter to link all: ")
	var input string

	fmt.Scanln(&input)

	if input == "" {
		for i := 1; i <= len(models); i++ {
			selectedModels = append(selectedModels, i)
		}
	} else {

		for _, numStr := range strings.Split(input, ",") {
			var num int
			fmt.Sscanf(numStr, "%d", &num)
			if num >= 1 && num <= len(models) {
				selectedModels = append(selectedModels, num)
			}
		}
	}

	for _, num := range selectedModels {
		modelName := models[num-1]
		modelPath, err := getModelPath(modelName)
		if err != nil {
			fmt.Printf("Error getting model path for %s: %v\n", modelName, err)
			continue
		}

		lmStudioModelName := strings.ReplaceAll(strings.ReplaceAll(modelName, ":", "-"), "_", "-")
		lmStudioModelDir := filepath.Join(lmStudioModelsDir, lmStudioModelName+"-GGUF")

		fmt.Printf("\033[1;36mModel:\033[0m %s\nPath: %s\n", modelName, modelPath)
		fmt.Printf("\033[1;32mLM Studio model directory:\033[0m %s\n", lmStudioModelDir)

		err = os.MkdirAll(lmStudioModelDir, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", lmStudioModelDir, err)
			continue
		}

		lmStudioModelPath := filepath.Join(lmStudioModelDir, lmStudioModelName+".gguf")

		// if the symlink already exists, delete it
		if _, err := os.Lstat(lmStudioModelPath); err == nil {
			fmt.Printf("Removing existing symlink: %s\n", lmStudioModelPath)
			err = os.Remove(lmStudioModelPath)
			if err != nil {
				fmt.Printf("Failed to remove symlink %s: %v\n", lmStudioModelPath, err)
				continue
			}
		}

		err = os.Symlink(modelPath, lmStudioModelPath)

		if err != nil {
			fmt.Printf("Failed to symlink %s: %v\n", modelName, err)
		} else {

			fmt.Printf("Symlinked %s to %s\n", modelName, lmStudioModelPath)
		}
	}

	cleanBrokenSymlinks()
}
