# Ollama to LM Studio Model Linker

This Go program is a command-line tool that allows you to easily link Ollama models to LM Studio's directory structure. It simplifies the process of symlinking Ollama models to LM Studio, making it convenient to use the models in both applications.

## Features

- Retrieves the list of available Ollama models using the ollama list command.
- Displays the list of models and allows the user to select specific models to link or link all models at once.
- Retrieves the model path for each selected model using the ollama show --modelfile command.
- Creates the necessary directories in the LM Studio directory structure.
- Creates symlinks from the Ollama model paths to the corresponding LM Studio model paths.
- Removes any existing symlinks before creating new ones to avoid conflicts.
- Cleans up any broken symlinks in the LM Studio models directory.

![screenshot](https://github.com/sammcj/llamalink/assets/862951/6559d22a-060f-42b9-9b31-e0c60f724d53)

## Prerequisites

Before using this program, ensure that you have the following:

- Go programming language installed on your system.
- Ollama command-line tool installed and accessible from the system PATH.
- Ollama models downloaded and stored in the default location (~/.ollama/models).

## Installation

- Clone the repository or download the source code files.
- Open a terminal and navigate to the directory containing the source code.
- Run the following command to build the program:

```shell
go build -o llamalink
```

This will create an executable file named ollama-lm-studio-linker (or ollama-lm-studio-linker.exe on Windows).

## Usage

- Open a terminal and navigate to the directory containing the executable file.
- Run the program using the following command:

```shell
./llamalink
```

Configuration

The program uses the following default directories:

- Ollama models directory: ~/.ollama/models
- LM Studio models directory: ~/.cache/lm-studio/models

If your Ollama models or LM Studio models are located in different directories, you can modify the ollamaModelsDir and lmStudioModelsDir variables in the source code accordingly.

## License

This program is open-source and available under the MIT License.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
