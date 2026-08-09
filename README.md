# agtgrep 🚀

**agtgrep** (Agent Grep) is a high-performance, structural code search engine designed specifically for AI agents, LLM tool-use pipelines, and developers who need surgical precision in their codebase.

Unlike traditional text-based search tools, `agtgrep` leverages **Tree-Sitter** to understand the abstract syntax tree (AST) of your code. This allows it to provide structured, context-rich results that are perfectly formatted for ingestion by Large Language Models (LLMs).

## ✨ Key Features

- **Structural AST Queries**: Use Tree-Sitter S-expressions (`-query`) to find specific code patterns (e.g., all function declarations, specific error returns, or class definitions).
- **Agent-Ready Output**: Native support for **JSON** and **XML** output formats, optimized for OpenAI function calling, Claude system prompts, and Gemini tool-use.
- **Multi-Language Support**: Built-in parsers for **Go, Python, JavaScript, TypeScript (TSX), and Rust**.
- **Context Awareness**: Automatically identifies the "parent block" (e.g., function signature or class name) for every match to provide the LLM with immediate architectural context.
- **Token Efficiency**: 
    - Respects `.gitignore` rules automatically.
    - Excludes binary files, build artifacts, and heavy directories like `node_modules` or `vendor`.
    - Truncates large code fragments to stay within token limits.
- **Fast & Portable**: Written in Go; ships as a single, static binary.

## 📦 Installation

### 1. One-Line Installer (Recommended)
Install `agtgrep` instantly:
```bash
curl -sSL [https://raw.githubusercontent.com/toshalkumbhar8979-design/Agent-grep/main/install.sh](https://raw.githubusercontent.com/toshalkumbhar8979-design/Agent-grep/main/install.sh) | sh
```

### 2. Using Go Install

If you have Go installed, you can install the latest version directly:

**Standard Install**
```bash
go install [github.com/toshalkumbhar8979-design/Agent-grep@latest](https://github.com/toshalkumbhar8979-design/Agent-grep@latest)
```
*Note: Ensure your `$GOPATH/bin` is in your `PATH`.*


## 🚀 Usage

### 1. Simple Symbol Search
Find all occurrences of a symbol (variable, function name, etc.) across the project:
```bash
agtgrep -symbol myFunctionName .
```

### 2. Structural Query (S-Expression)
Find all function declarations in your Go project:
```bash
agtgrep -query "(function_declaration) @func" .
```

### 3. XML Output for Claude/Gemini
Output results in XML format, which is often preferred by certain LLM providers for clear block boundaries:
```bash
agtgrep -format xml -symbol Config .
```

### 4. Excluding Custom Patterns
Ignore specific files or directories that aren't already in your `.gitignore`:
```bash
agtgrep -exclude "*.test.go,docs/*" -symbol Run .
```

### 5. Filter by AST Node Type
Search for a symbol but only where it acts as a specific node type:
```bash
agtgrep -symbol os -type identifier .
```

## 🛠 Command Line Flags

| Flag | Description | Default |
| :--- | :--- | :--- |
| `-dir` | Target directory to search | `.` |
| `-symbol` | Target symbol to search for | (Required) |
| `-query` | Custom Tree-Sitter S-expression query | `""` |
| `-format` | Output format (`json` or `xml`) | `json` |
| `-type` | Filter by specific AST node type | `""` |
| `-exclude` | Comma-separated glob patterns to ignore | `""` |
| `-max-tokens` | Approx max tokens for code fragment | `500` |

## 🌐 Supported Languages

- **Go** (`.go`)
- **Python** (`.py`)
- **JavaScript** (`.js`)
- **TypeScript** (`.ts`, `.tsx`)
- **Rust** (`.rs`)

---

## 📄 License
MIT

## 🤝 Contributing
Contributions are welcome! Please feel free to submit a Pull Request.
