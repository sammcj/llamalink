package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var ollamaModelsDir string
var lmStudioModelsDir string

// define output colours
const (
	red          = "\033[1;31m"
	green        = "\033[1;32m"
	yellow       = "\033[1;33m"
	brightYellow = "\033[1;93m"
	blue         = "\033[1;34m"
	purple       = "\033[1;35m"
	cyan         = "\033[1;36m"
	white        = "\033[1;37m"
	brightWhite  = "\033[1;97m"
	grey         = "\033[1;30m"
	reset        = "\033[0m"
)

func printHelp() {
	fmt.Printf("%sUsage: ollama-lm-studio-linker [options]%s\n", brightWhite, reset)
	fmt.Printf("%sOptions:%s\n", brightWhite, reset)
	fmt.Printf("  %s-a%s           Link all available models\n", yellow, reset)
	fmt.Printf("  %s-h%s           Print this help message\n", yellow, reset)
	fmt.Printf("  %s-ollama-dir%s  Custom Ollama models directory\n", yellow, reset)
	fmt.Printf("  %s-lm-dir%s      Custom LM Studio models directory\n", yellow, reset)
	fmt.Printf("  %s-min-size%s    Include only models over the given size (in GB or MB)\n", yellow, reset)
	fmt.Printf("  %s-max-size%s    Include only models under the given size (in GB or MB)\n", yellow, reset)
	fmt.Printf("  %s-q%s           Quiet operation, only output an exit code at the end\n", yellow, reset)
	fmt.Printf("  %s-no-cleanup%s  Don't cleanup broken symlinks\n", yellow, reset)
	fmt.Printf("  %s-cleanup%s     Remove all symlinked models and empty directories and exit\n", yellow, reset)
}

// print the configured model paths
func printModelPaths() {
	fmt.Printf("%sOllama models directory:%s %s\n", brightWhite, reset, ollamaModelsDir)
	fmt.Printf("%sLM Studio models directory:%s %s\n", brightWhite, reset, lmStudioModelsDir)
	fmt.Println()
}

func getModelList(minSize, maxSize float64) ([]string, error) {
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
			modelName := fields[0]
			if len(fields) > 1 {
				sizeStr := fields[1]
				sizeFloat, err := strconv.ParseFloat(sizeStr[:len(sizeStr)-2], 64)
				if err != nil {
					// If the size is not available or cannot be parsed, consider it as 0
					sizeFloat = 0
				}
				if strings.HasSuffix(sizeStr, "MB") {
					sizeFloat /= 1024
				}
				if sizeFloat >= minSize && (maxSize == 0 || sizeFloat <= maxSize) {
					models = append(models, modelName)
				}
			} else {
				// If the size is not available, include the model if no size range is specified
				if minSize == 0 && maxSize == 0 {
					models = append(models, modelName)
				}
			}
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
		if info.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Printf("%sRemoving empty directory: %s%s\n", yellow, path, reset)
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}
		} else if info.Mode()&os.ModeSymlink != 0 {
			linkPath, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if !isValidSymlink(path, linkPath) {
				fmt.Printf("%sRemoving invalid symlink: %s%s\n", yellow, path, reset)
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("%sError walking LM Studio models directory: %v%s\n", red, err, reset)
	}
}

func isValidSymlink(symlinkPath, targetPath string) bool {
	// Check if the symlink matches the expected naming convention
	expectedSuffix := ".gguf"
	if !strings.HasSuffix(filepath.Base(symlinkPath), expectedSuffix) {
		return false
	}

	// Check if the target file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return false
	}

	// Check if the symlink target is a file (not a directory or another symlink)
	fileInfo, err := os.Lstat(targetPath)
	if err != nil || fileInfo.Mode()&os.ModeSymlink != 0 || fileInfo.IsDir() {
		return false
	}

	return true
}

func cleanupSymlinkedModels() {
	for {
		hasEmptyDir := false
		err := filepath.Walk(lmStudioModelsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				files, err := os.ReadDir(path)
				if err != nil {
					return err
				}
				if len(files) == 0 {
					fmt.Printf("%sRemoving empty directory: %s%s\n", yellow, path, reset)
					err = os.Remove(path)
					if err != nil {
						return err
					}
					hasEmptyDir = true
				}
			} else if info.Mode()&os.ModeSymlink != 0 {
				fmt.Printf("%sRemoving symlinked model: %s%s\n", yellow, path, reset)
				err = os.Remove(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("%sError walking LM Studio models directory: %v%s\n", red, err, reset)
			return
		}
		if !hasEmptyDir {
			break
		}
	}
}

func main() {
	// Parse command-line arguments
	linkAllModels := flag.Bool("a", false, "Link all available models")
	printHelpFlag := flag.Bool("h", false, "Print help message")
	ollamaDirFlag := flag.String("ollama-dir", "", "Custom Ollama models directory")
	lmStudioDirFlag := flag.String("lm-dir", "", "Custom LM Studio models directory")
	minSizeFlag := flag.Float64("min-size", 0, "Include only models over the given size (in GB or MB)")
	maxSizeFlag := flag.Float64("max-size", 0, "Include only models under the given size (in GB or MB)")
	quietFlag := flag.Bool("q", false, "Quiet operation, only output an exit code at the end")
	noCleanupFlag := flag.Bool("no-cleanup", false, "Don't cleanup broken symlinks")
	cleanupFlag := flag.Bool("cleanup", false, "Remove all symlinked models and empty directories and exit")
	flag.Parse()

	// Print help if -h flag is provided
	if *printHelpFlag {
		printHelp()
		return
	}

	// Set custom model directories if provided
	if *ollamaDirFlag != "" {
		ollamaModelsDir = *ollamaDirFlag
	} else {
		ollamaModelsDir = filepath.Join(os.Getenv("HOME"), ".ollama", "models")
	}
	if *lmStudioDirFlag != "" {
		lmStudioModelsDir = *lmStudioDirFlag
	} else {
		lmStudioModelsDir = filepath.Join(os.Getenv("HOME"), ".cache", "lm-studio", "models")
	}

	if *cleanupFlag {
		cleanupSymlinkedModels()
		os.Exit(0)
	}

	if !*quietFlag {
		printModelPaths()
	}

	models, err := getModelList(*minSizeFlag, *maxSizeFlag)
	if err != nil {
		if !*quietFlag {
			fmt.Printf("%sError getting model list: %v%s\n", red, err, reset)
		}
		os.Exit(1)
	}

	if len(models) == 0 {
		if !*quietFlag {
			fmt.Printf("%sNo Ollama models found.%s\n", yellow, reset)
		}
		os.Exit(0)
	}

	if !*quietFlag {
		fmt.Printf("%sSelect the models to link to LM Studio:%s\n", brightWhite, reset)

		for i, modelName := range models {
			fmt.Printf("%s%d.%s %s\n", brightYellow, i+1, reset, modelName)
		}
	}

	var selectedModels []int

	// If -a flag is provided, link all models
	if *linkAllModels {
		for i := 1; i <= len(models); i++ {
			selectedModels = append(selectedModels, i)
		}
	} else if !*quietFlag {
		fmt.Println()
		fmt.Printf("%sEnter the model numbers (comma-separated), or press Enter to link all: %s", brightWhite, reset)
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
	}

	for _, num := range selectedModels {
		modelName := models[num-1]
		modelPath, err := getModelPath(modelName)
		if err != nil {
			if !*quietFlag {
				fmt.Printf("%sError getting model path for %s: %v%s\n", red, modelName, err, reset)
			}
			continue
		}

		parts := strings.Split(modelName, ":")
		author := "unknown"
		if len(parts) > 1 {
			author = strings.ReplaceAll(parts[0], "/", "-")
		}

		lmStudioModelName := strings.ReplaceAll(strings.ReplaceAll(modelName, ":", "-"), "_", "-")
		lmStudioModelDir := filepath.Join(lmStudioModelsDir, author, lmStudioModelName+"-GGUF")

		if !*quietFlag {
			fmt.Printf("%sModel:%s %s\nPath: %s\n", cyan, reset, modelName, modelPath)
			fmt.Printf("%sLM Studio model directory:%s %s\n", green, reset, lmStudioModelDir)
		}

		// Check if the model path is a valid file
		fileInfo, err := os.Stat(modelPath)
		if err != nil || fileInfo.IsDir() {
			if !*quietFlag {
				fmt.Printf("%sInvalid model path for %s: %s%s\n", red, modelName, modelPath, reset)
			}
			continue
		}

		// Check if the symlink already exists and is valid
		lmStudioModelPath := filepath.Join(lmStudioModelDir, filepath.Base(lmStudioModelName)+".gguf")
		if _, err := os.Lstat(lmStudioModelPath); err == nil {
			if isValidSymlink(lmStudioModelPath, modelPath) {
				if !*quietFlag {
					fmt.Printf("%sFound model %s, already linked, skipping...%s\n", grey, modelName, reset)
				}
				continue
			}
			// Remove the invalid symlink
			err = os.Remove(lmStudioModelPath)
			if err != nil {
				if !*quietFlag {
					fmt.Printf("%sFailed to remove invalid symlink %s: %v%s\n", red, lmStudioModelPath, err, reset)
				}
			}
		}

		// Check if the model is already symlinked in another location
		var existingSymlinkPath string
		err = filepath.Walk(lmStudioModelsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				linkPath, err := os.Readlink(path)
				if err != nil {
					return err
				}
				if linkPath == modelPath {
					existingSymlinkPath = path
					return nil
				}
			}
			return nil
		})
		if err != nil {
			if !*quietFlag {
				fmt.Printf("%sError checking for duplicated symlinks: %v%s\n", red, err, reset)
			}
			continue
		}

		if existingSymlinkPath != "" {
			// Remove the duplicated model directory
			err = os.RemoveAll(lmStudioModelDir)
			if err != nil {
				if !*quietFlag {
					fmt.Printf("%sFailed to remove duplicated model directory %s: %v%s\n", red, lmStudioModelDir, err, reset)
				}
			} else {
				if !*quietFlag {
					fmt.Printf("%sRemoved duplicated model directory %s%s\n", yellow, lmStudioModelDir, reset)
				}
			}
			continue
		}

		// Create the symlink
		err = os.MkdirAll(lmStudioModelDir, os.ModePerm)
		if err != nil {
			if !*quietFlag {
				fmt.Printf("%sFailed to create directory %s: %v%s\n", red, lmStudioModelDir, err, reset)
			}
			continue
		}
		err = os.Symlink(modelPath, lmStudioModelPath)
		if err != nil {
			if !*quietFlag {
				fmt.Printf("%sFailed to symlink %s: %v%s\n", red, modelName, err, reset)
			}
		} else {
			if !*quietFlag {
				fmt.Printf("%sSymlinked %s to %s%s\n", green, modelName, lmStudioModelPath, reset)
			}

			if !*quietFlag {
				fmt.Println()
			}
		}

		if !*noCleanupFlag {
			cleanBrokenSymlinks()
		}
	}

	os.Exit(0)
}
