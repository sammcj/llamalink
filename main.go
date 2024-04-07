package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type LlamaLinkConfig struct {
	LastSyncTime           int64 `json:"lastSyncTime"`
	SkipConfigPresets      bool  `json:"skipConfigPresets"`
	OverwriteConfigPresets bool  `json:"overwriteConfigPresets"`
	SkipModelfileSync      bool  `json:"skipModelfileSync"`
	ReportOnly             bool  `json:"reportOnly"`
}

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
	fmt.Printf("%sUsage: llamalink [options]%s\n", brightWhite, reset)
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
	fmt.Printf("  %s-skip-config-presets%s  Skip syncing config presets\n", yellow, reset)
	fmt.Printf("  %s-overwrite-config-presets%s  Overwrite existing config presets\n", yellow, reset)
	fmt.Printf("  %s-skip-modelfile-sync%s  Skip syncing modelfiles\n", yellow, reset)
	fmt.Printf("  %s-report-only%s  Generate a report without creating symlinks or manifests\n", yellow, reset)
	fmt.Printf("  %s-show-config-preset%s  Show the config preset for a specific model\n", yellow, reset)
}

func loadLlamaLinkConfig() (*LlamaLinkConfig, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "llamalink")
	configFile := filepath.Join(configDir, "config.json")

	// Create the config directory if it doesn't exist
	err := os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If the config file doesn't exist, return a default config
			return &LlamaLinkConfig{}, nil
		}
		return nil, err
	}

	// Parse the config JSON
	var config LlamaLinkConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func saveLlamaLinkConfig(config *LlamaLinkConfig) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "llamalink")
	configFile := filepath.Join(configDir, "config.json")

	// Marshal the config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write the config file
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func syncConfigPresets(models []string, config *LlamaLinkConfig) error {
	if config.SkipConfigPresets {
		log.Println("Skipping config preset sync")
		return nil
	}

	configPresetsDir := filepath.Join(os.Getenv("HOME"), ".cache", "lm-studio", "config-presets")
	configMapFile := filepath.Join(configPresetsDir, "config.map.json")

	// Read the existing config map
	configMapData, err := os.ReadFile(configMapFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read config map file: %v", err)
		}
		configMapData = []byte("{}")
	}

	var configMap map[string]interface{}
	err = json.Unmarshal(configMapData, &configMap)
	if err != nil {
		return fmt.Errorf("failed to parse config map JSON: %v", err)
	}

	// Update the config map with the correct preset mappings
	for _, modelName := range models {
		// Determine the correct preset mapping based on the model name
		var presetFile string
		// ... (logic to determine the preset file based on the model name)

		// Example logic: Use a specific preset file for models containing "codellama"
		if strings.Contains(modelName, "codellama") {
			presetFile = "codellama_instruct.preset.json"
		}

		if presetFile != "" {
			// Check if the preset file exists
			presetFilePath := filepath.Join(configPresetsDir, presetFile)
			if _, err := os.Stat(presetFilePath); os.IsNotExist(err) {
				// Create the preset file if it doesn't exist
				err = createConfigPresetFile(presetFilePath)
				if err != nil {
					log.Printf("Failed to create config preset file %s: %v", presetFile, err)
					continue
				}
			}

			// Update the config map
			if configMap["preset_map"] == nil {
				configMap["preset_map"] = make(map[string]interface{})
			}
			configMap["preset_map"].(map[string]interface{})[modelName] = presetFile
		}
	}

	// Write the updated config map
	configMapData, err = json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config map JSON: %v", err)
	}

	err = os.WriteFile(configMapFile, configMapData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config map file: %v", err)
	}

	log.Println("Config presets synced successfully")
	return nil
}

func createConfigPresetFile(presetFilePath string) error {
	// Create the config preset file with a default template
	defaultPresetData := []byte(`{
  "name": "Default Preset",
  "inference_params": {
    "input_prefix": "### Instruction:\\n",
    "input_suffix": "\\n### Response:\\n",
    "antiprompt": [
      "### Instruction:"
    ],
    "pre_prompt": "Below is an instruction that describes a task. Write a response that appropriately completes the request.",
    "pre_prompt_suffix": "\\n",
    "pre_prompt_prefix": ""
  }
}`)

	err := os.WriteFile(presetFilePath, defaultPresetData, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config preset file: %v", err)
	}

	return nil
}

func syncModelfiles(config *LlamaLinkConfig) error {
	if config.SkipModelfileSync {
		log.Println("Skipping modelfile sync")
		return nil
	}

	// Get the list of model folders in LM Studio
	lmStudioModelFolders, err := os.ReadDir(lmStudioModelsDir)
	if err != nil {
		return fmt.Errorf("failed to read LM Studio models directory: %v", err)
	}

	// Iterate over the model folders
	for _, modelFolder := range lmStudioModelFolders {
		if !modelFolder.IsDir() {
			continue
		}

		modelFolderPath := filepath.Join(lmStudioModelsDir, modelFolder.Name())
		modelFiles, err := os.ReadDir(modelFolderPath)
		if err != nil {
			log.Printf("Failed to read model folder %s: %v", modelFolderPath, err)
			continue
		}

		// Find the model file (*.gguf)
		var modelFilePath string
		for _, modelFile := range modelFiles {
			if !modelFile.IsDir() && strings.HasSuffix(modelFile.Name(), ".gguf") {
				modelFilePath = filepath.Join(modelFolderPath, modelFile.Name())
				break
			}
		}

		if modelFilePath == "" {
			// log.Printf("No model file found in folder %s", modelFolderPath)
			continue
		}

		// Determine the model name from the folder name
		modelName := strings.TrimSuffix(modelFolder.Name(), "-GGUF")
		modelName = strings.ReplaceAll(modelName, "-", ":")
		modelName = strings.ReplaceAll(modelName, "_", ":")

		// Check if the model is already symlinked in Ollama
		ollamaModelPath := filepath.Join(ollamaModelsDir, "blobs", "sha256-model-"+modelName)
		if _, err := os.Lstat(ollamaModelPath); err == nil {
			// Model is already symlinked, skip
			continue
		}

		// Create the symlink
		err = os.Symlink(modelFilePath, ollamaModelPath)
		if err != nil {
			log.Printf("Failed to symlink model %s: %v", modelName, err)
			continue
		}

		// Create the Ollama modelfile manifest
		manifestPath := filepath.Join(ollamaModelsDir, "manifests", "registry.ollama.ai", strings.Replace(modelName, ":", "/", -1))
		manifestData := []byte(fmt.Sprintf(`{
			"schemaVersion": 2,
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"config": {
				"mediaType": "application/vnd.docker.container.image.v1+json",
				"digest": "sha256:%s",
				"size": %d
			},
			"layers": [
				{
					"mediaType": "application/vnd.ollama.image.model",
					"digest": "sha256:%s",
					"size": %d
				}
			]
		}`, modelName, os.Getpagesize(), modelName, os.Getpagesize()))

		err = os.WriteFile(manifestPath, manifestData, 0644)
		if err != nil {
			log.Printf("Failed to create modelfile manifest for %s: %v", modelName, err)
			continue
		}
	}

	log.Println("Modelfiles synced successfully")
	return nil
}

func generateReport(config *LlamaLinkConfig) error {
	if !config.ReportOnly {
		return nil
	}

	// Check for missing symlinks in LM Studio
	lmStudioMissingSymlinks := []string{}
	err := filepath.Walk(lmStudioModelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".gguf") {
			return nil
		}
		if _, err := os.Readlink(path); err != nil {
			lmStudioMissingSymlinks = append(lmStudioMissingSymlinks, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk LM Studio models directory: %v", err)
	}

	// Check for missing modelfile manifests in Ollama
	ollamaMissingManifests := []string{}
	err = filepath.Walk(ollamaModelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasPrefix(info.Name(), "sha256-model-") {
			return nil
		}
		modelName := strings.TrimPrefix(info.Name(), "sha256-model-")
		manifestPath := filepath.Join(ollamaModelsDir, "manifests", "registry.ollama.ai", strings.Replace(modelName, ":", "/", -1))
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			ollamaMissingManifests = append(ollamaMissingManifests, modelName)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk Ollama models directory: %v", err)
	}

	// Generate the report
	fmt.Printf("Missing symlinks in LM Studio:\n")
	for _, symlink := range lmStudioMissingSymlinks {
		fmt.Printf("- %s\n", symlink)
	}
	fmt.Printf("\nMissing modelfile manifests in Ollama:\n")
	for _, manifest := range ollamaMissingManifests {
		fmt.Printf("- %s\n", manifest)
	}

	log.Println("Report generated successfully")
	return nil
}

func showConfigPreset(modelName string) error {
	configPresetsDir := filepath.Join(os.Getenv("HOME"), ".cache", "lm-studio", "config-presets")
	configMapFile := filepath.Join(configPresetsDir, "config.map.json")

	// Read the config map
	configMapData, err := os.ReadFile(configMapFile)
	if err != nil {
		return err
	}

	var configMap map[string]interface{}
	err = json.Unmarshal(configMapData, &configMap)
	if err != nil {
		return err
	}

	// Find the config preset for the model
	var presetFile string
	for pattern, preset := range configMap["preset_map"].(map[string]interface{}) {
		if matched, _ := regexp.MatchString(pattern, modelName); matched {
			presetFile = preset.(string)
			break
		}
	}

	if presetFile == "" {
		fmt.Printf("No config preset found for model: %s\n", modelName)
		return nil
	}

	// Read the config preset file
	presetData, err := os.ReadFile(filepath.Join(configPresetsDir, presetFile))
	if err != nil {
		return err
	}

	// Print the config preset
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, presetData, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(prettyJSON.String())
	return nil
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
	log.Printf("Executing command: %s", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
			log.Printf("Error running ollama show command for model %s: %v", modelName, err)
			log.Printf("Command output: %s", string(output))
			return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "FROM ") {
			modelPath := strings.TrimSpace(line[5:])
			log.Printf("Model path for %s: %s", modelName, modelPath)
			return modelPath, nil
		}
	}
	log.Printf("Model path not found for %s", modelName)
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

	// Check if the target file exists
	_, err = os.Stat(targetPath)
	if os.IsNotExist(err) {
		return false
	}

	// Check if the symlink target is a file (not a directory or another symlink)
	fileInfo, err = os.Lstat(targetPath)
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
	skipConfigPresetsFlag := flag.Bool("skip-config-presets", false, "Skip syncing config presets")
	overwriteConfigPresetsFlag := flag.Bool("overwrite-config-presets", false, "Overwrite existing config presets")
	skipModelfileSyncFlag := flag.Bool("skip-modelfile-sync", false, "Skip syncing modelfiles")
	reportOnlyFlag := flag.Bool("report-only", false, "Generate a report without creating symlinks or manifests")
	showConfigPresetFlag := flag.String("show-config-preset", "", "Show the config preset for a specific model")
	flag.Parse()
	// Print help if -h flag is provided
	if *printHelpFlag {
		printHelp()
		return
	}

	// Load the LlamaLink config
	config, err := loadLlamaLinkConfig()
	if err != nil {
		log.Fatalf("Error loading LlamaLink config: %v", err)
	}

	// Update the config based on command-line flags
	config.SkipConfigPresets = *skipConfigPresetsFlag
	config.OverwriteConfigPresets = *overwriteConfigPresetsFlag
	config.SkipModelfileSync = *skipModelfileSyncFlag
	config.ReportOnly = *reportOnlyFlag

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

	// Show config preset
	if *showConfigPresetFlag != "" {
		err := showConfigPreset(*showConfigPresetFlag)
		if err != nil {
			log.Fatalf("Error showing config preset: %v", err)
		}
		os.Exit(0)
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

	// Sync config presets
	err = syncConfigPresets(models, config)
	if err != nil {
		log.Fatalf("Error syncing config presets: %v", err)
	}

	// Sync modelfiles
	err = syncModelfiles(config)
	if err != nil {
		log.Fatalf("Error syncing modelfiles: %v", err)
	}

	// Generate report
	err = generateReport(config)
	if err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	for _, num := range selectedModels {
		modelName := models[num-1]
		log.Printf("Selected model: %s", modelName)
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

	// Save the updated LlamaLink config
	err = saveLlamaLinkConfig(config)
	if err != nil {
		log.Fatalf("Error saving LlamaLink config: %v", err)
	}

	os.Exit(0)
}
