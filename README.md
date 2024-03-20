# Ollama to LM Studio Model Linker

This is a simple command-line tool that allows you to easily link Ollama models to LM Studio's directory structure. It simplifies the process of symlinking Ollama models to LM Studio, making it convenient to use the models in both applications.

## Features

- Retrieves the list of available Ollama models using the ollama list command.
- Displays the list of models and allows the user to select specific models to link or link all models at once.
- Retrieves the model path for each selected model using the ollama show --modelfile command.
- Creates the necessary directories in the LM Studio directory structure.
- Creates symlinks from the Ollama model paths to the corresponding LM Studio model paths.
- Removes any existing symlinks before creating new ones to avoid conflicts.
- Cleans up any broken symlinks in the LM Studio models directory.
- Can be run interactively or non-interactively.

![screenshot](https://github.com/sammcj/llamalink/assets/862951/6559d22a-060f-42b9-9b31-e0c60f724d53)

## Prerequisites

Before using this program, ensure that you have the following:

- Ollama command-line tool installed and accessible from the system PATH.

## Installation

- Run `go install github.com/sammcj/llamalink@latest`

or

- Download the latest build from the [releases page](https://github.com/sammcj/llamalink/releases).

## Usage

- Open a terminal and navigate to the directory containing the executable file.
- Run the program using the following command:

```shell
./llamalink
```

or non-interactively, linking all models:

```shell
./llamalink -a
```

Configuration

The program uses the following default directories:

- Ollama models directory: ~/.ollama/models
- LM Studio models directory: ~/.cache/lm-studio/models

If your Ollama models or LM Studio models are located in different directories, you can modify the ollamaModelsDir and lmStudioModelsDir variables in the source code accordingly.

```plaintext
$ llamalink

Ollama models directory: /Users/samm/.ollama/models
LM Studio models directory: /Users/samm/.cache/lm-studio/models

Select the models to link to LM Studio:
1. knoopx/hermes-2-pro-mistral:7b-q8_0
2. dolphincoder:15b-starcoder2-q4_K_M
3. llava:13b
4. nomic-embed-text:latest
5. qwen:14b
6. qwen:7b-chat-q5_K_M
7. stable-code:3b-code-q5_K_M
8. tinydolphin:1.1b-v2.8-q5_K_M

Enter the model numbers (comma-separated), or press Enter to link all: 1
Model: knoopx/hermes-2-pro-mistral:7b-q8_0
Path: /Users/samm/.ollama/models/blobs/sha256:107d9516acb6a1f879b1fbfa283b399529ee0518b95b632c6a624b109ff9cdbf
LM Studio model directory: /Users/samm/.cache/lm-studio/models/knoopx/hermes-2-pro-mistral-7b-q8-0-GGUF
Removing existing symlink: /Users/samm/.cache/lm-studio/models/knoopx/hermes-2-pro-mistral-7b-q8-0-GGUF/sha256:107d9516acb6a1f879b1fbfa283b399529ee0518b95b632c6a624b109ff9cdbf
Symlinked knoopx/hermes-2-pro-mistral:7b-q8_0 to /Users/samm/.cache/lm-studio/models/knoopx/hermes-2-pro-mistral-7b-q8-0-GGUF/sha256:107d9516acb6a1f879b1fbfa283b399529ee0518b95b632c6a624b109ff9cdbf
```

## Building

```shell
go build
```

## License

This program is open-source and available under the MIT License.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
