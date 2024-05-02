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

![screenshot](https://github.com/sammcj/llamalink/assets/862951/0131f98a-9940-434b-abcf-2594ab09709c)

## Prerequisites

Before using this program, ensure that you have the following:

- Ollama command-line tool installed and accessible from the system PATH.

## Installation

- Run `go install github.com/sammcj/llamalink@latest`

or

- Download the latest build from the [releases page](https://github.com/sammcj/llamalink/releases).

## Usage

Install the program using the instructions above.

Run `llamalink`

- `-i` Run interactively to select which specific models to link.
- `-ollama-dir` Specify a custom Ollama models directory.
- `-lm-dir` Specify a custom LM Studio models directory.
- `-min-size` Include only models over the given size (in GB or MB).
- `-max-size` Include only models under the given size (in GB or MB).
- `-q` Quiet operation, only output an exit code at the end.
- `-no-cleanup` Don't cleanup broken symlinks.
- `-cleanup` Remove all symlinked models and empty directories and exit.
- `-h` Print the help message.

If no flags are provided, the program will automatically link all models.

## Configuration

The program uses the following default directories:

- Ollama models directory: ~/.ollama/models
- LM Studio models directory: ~/.cache/lm-studio/models

If your Ollama models or LM Studio models are located in different directories, you can modify the ollamaModelsDir and lmStudioModelsDir variables in the source code accordingly.

```plaintext
$ go install github.com/sammcj/llamalink@latest
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

Enter the model numbers (comma-separated), or press Enter to link all: 3
Model: llava:34b-v1.6-q5_K_M
Path: /Users/samm/.ollama/models/blobs/sha256:4ddc7cc65a231db765ecf716d58ed3262e4496847eafcbcf80288fc2c552d9e6
LM Studio model directory: /Users/samm/.cache/lm-studio/models/llava/llava-34b-v1.6-q5-K-M-GGUF
Symlinked llava:34b-v1.6-q5_K_M to /Users/samm/.cache/lm-studio/models/llava/llava-34b-v1.6-q5-K-M-GGUF/llava-34b-v1.6-q5-K-M.gguf
```

## Building

```shell
go build
```

## License

This program is open-source and available under the MIT License.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
