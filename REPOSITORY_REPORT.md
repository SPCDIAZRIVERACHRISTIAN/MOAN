# MOAN Repository Report
**Generated:** 2026-06-12

---

## Project Overview

**Project Name:** MOAN (Modular Orchestration for Automated Nodes)

**Type:** Go CLI Application

**Purpose:** A local-first code review tool that analyzes Git changes and sends them to AI models (primarily Ollama) for automated code review with a focus on bugs, security, architecture, and maintainability issues.

**Module:** `github.com/SPCDIAZRIVERACHRISTIAN/moan`

**Go Version:** 1.26.1

---

## Current Status

**Stage:** Early MVP Development

**Branch:** MVP (diverged from main)

**Last Commit:** `f66e10d` - "last human produced code"

---

## Repository Structure

```
MOAN/
├── cmd/                          # CLI commands
│   ├── config.go                 # Configuration command
│   ├── review.go                 # Review command (main entry point)
│   ├── root.go                   # Root command definition
│   ├── validate.go               # Validation command
│   └── version.go                # Version command
├── internal/                     # Core application logic
│   ├── config/
│   │   └── config.go             # Configuration loading and management
│   ├── git/
│   │   ├── diff.go               # Git diff parsing and extraction
│   │   └── state.go              # Git repository state validation
│   ├── provider/
│   │   ├── factory.go            # Provider factory pattern
│   │   ├── ollama.go             # Ollama provider implementation
│   │   └── provider.go           # Provider interface definition
│   ├── review/
│   │   └── review.go             # Core review orchestration logic
│   └── validate/
│       └── validate.go           # Repository validation logic
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── LICENSE                       # License file
└── README.md                     # Project README
```

---

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/spf13/cobra` | v1.10.2 | CLI framework |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML configuration parsing |
| `github.com/inconshreveable/mousetrap` | v1.1.0 | Windows CLI utility (indirect) |
| `github.com/spf13/pflag` | v1.0.9 | Flag parsing (indirect) |

---

## Core Components

### 1. **CLI Interface** (`cmd/`)

#### Root Command (`root.go`)
- Base Cobra command definition
- Use: `moan`
- Description: "moan is a code reviewer orchestrator"
- Long: "MOAN analyzes code changes and runs structured review passes for bugs, security, and architecture."
- Unused toggle flag present (marked for cleanup)

#### Review Command (`review.go`)
- **Primary entry point** for code review workflow
- Features:
  - Animated loader showing review progress (`[m   ] → [MOAN]`)
  - Validates repository state before review
  - Loads configuration
  - Gathers changed file statistics
  - Sends diff to configured AI model
  - Formats and displays results

- Output Format:
  ```
  STATUS: READY/NOT READY
  Provider: <provider_name>
  Model: <model_name>
  Files changed: <count>
  - <file> | +<additions> -<deletions>
  
  AI REVIEW
  ---------
  <model_response>
  ```

#### Other Commands
- **validate.go**: Validates repository prerequisites
- **config.go**: Configuration management (implementation details not fully shown)
- **version.go**: Version information (implementation details not fully shown)

---

### 2. **Review Logic** (`internal/review/`)

#### Core Orchestration (`review.go`)
- **Main Review Pipeline:**
  1. Validates repository is Git and has changes
  2. Loads configuration (provider, model, system prompt)
  3. Gathers changed file statistics via Git
  4. Creates ReviewResult with file metadata
  5. Initializes provider (Ollama)
  6. Retrieves full Git diff content
  7. Builds detailed review prompt
  8. Sends request to AI model with 90-second timeout
  9. Returns formatted results

- **Data Structures:**
  - `FileChange`: Tracks path, additions, deletions per file
  - `ReviewResult`: Complete review outcome with provider info and AI response

- **Prompt Engineering:**
  - Strict, high-quality review criteria embedded in system prompt
  - Focuses on concrete bugs, security risks, reliability issues (not style)
  - Clear finding categories: bug, security, reliability, performance, maintainability, test, documentation
  - Severity levels: critical, high, medium, low, info
  - Enforces evidence-based findings with specific fixes

---

### 3. **Git Integration** (`internal/git/`)

#### Diff Extraction (`diff.go`)
- **GetChangedFileStats()**: Returns aggregated statistics for all changed files
  - Combines unstaged and staged changes
  - Returns: file path, additions count, deletions count
  - Handles binary files ("-" values)

- **GetDiffContent()**: Retrieves full Git diff
  - Separates unstaged and staged changes in output
  - Includes complete diff context for AI review

#### State Management (`state.go`)
- Validates repository prerequisites
- Checks for Git repository
- Verifies changes exist

---

### 4. **Provider System** (`internal/provider/`)

#### Provider Interface (`provider.go`)
- Abstraction for different AI backends
- Methods:
  - `Review()`: Submit review request and get response
  - `TestConnection()`: Verify provider connectivity

- Request/Response structures:
  - `ReviewRequest`: model, system prompt, user messages
  - `ReviewResponse`: content, raw response
  - `Message`: role-based message format

#### Ollama Implementation (`ollama.go`)
- **HTTP-based Ollama API client**
- Features:
  - Configurable base URL
  - 600-second timeout for long-running models
  - Connection testing via `/api/tags` endpoint
  - Chat completion via `/api/chat` endpoint
  - Structured JSON request/response handling

- **Request Format:**
  ```json
  {
    "model": "model_name",
    "stream": false,
    "messages": [
      {"role": "system", "content": "system_prompt"},
      {"role": "user", "content": "review_request"}
    ]
  }
  ```

#### Provider Factory (`factory.go`)
- Instantiates appropriate provider based on configuration
- Currently supports: Ollama

---

### 5. **Configuration System** (`internal/config/`)

#### Config Management (`config.go`)
- Loads configuration from:
  - Environment or default config file
- Manages:
  - Provider type
  - Model name
  - System prompt for review agent
  - API endpoints

---

### 6. **Validation System** (`internal/validate/`)

#### Pre-flight Checks (`validate.go`)
- Ensures Git repository exists
- Verifies Ollama connectivity
- Validates configuration
- Returns validation status

---

## Current Capabilities

✅ **Implemented:**
- Git repository validation
- Detection of tracked file changes
- Reading changed file metadata (stats)
- Connection to local Ollama models
- Sending review requests to configured models
- Returning AI-generated review feedback
- Detailed prompt engineering for high-quality reviews
- Support for both staged and unstaged changes
- File statistics display (additions/deletions)

---

## Planned Capabilities

📋 **Not Yet Implemented:**
- Full interactive config command (UI/prompts)
- Multi-agent review system
- Specialized security reviewer agent
- Specialized architecture reviewer agent
- Specialized test/QA reviewer agent
- Review report generation/persistence
- Commit message generation
- Human approval workflow before pushing
- Support for multiple AI providers (OpenAI, Anthropic, etc.)
- Review output formatting options (JSON, HTML, etc.)

---

## System Requirements

**Required:**
- Go 1.26.1+
- Git
- Ollama (running locally)
- A local Ollama model (recommended: `qwen2.5-coder:7b`)

**Optional:**
- YAML config file for persistent settings

---

## Review Quality Features

The system enforces strict review standards:

1. **Precision-First Approach**
   - Only report concrete, evidence-based findings
   - No generic best-practice advice
   - No style/formatting complaints

2. **Valid Finding Requirements**
   - Must cite specific file from diff
   - Must describe concrete issue
   - Must provide evidence from diff
   - Must explain realistic impact
   - Must suggest actionable fix

3. **Invalid Finding Types (Explicitly Rejected)**
   - Style preferences
   - Verbosity concerns
   - Naming conventions
   - Hypothetical architecture
   - Assumptions about unseen code
   - General advice unrelated to diff

4. **Severity Ranking**
   - **Critical**: Breaks functionality, data loss, severe security risk
   - **High**: Likely bug, security issue, major broken behavior
   - **Medium**: Real issue with meaningful impact
   - **Low**: Minor but concrete issue
   - **Info**: Useful notes tied to diff

---

## Code Quality Observations

### Strengths
- ✓ Clean separation of concerns (CLI, review logic, providers, Git)
- ✓ Well-designed abstractions (provider interface pattern)
- ✓ Type-safe configuration and data structures
- ✓ Context-aware API calls with timeouts
- ✓ Comprehensive prompt engineering
- ✓ Proper error handling with context

### Areas for Enhancement
- Review command could have flags for custom models/providers
- Configuration command implementation needed (placeholder only)
- Version command implementation details not visible
- No tests visible (TDD opportunities)
- Hard-coded review timeout (90s) could be configurable
- No logging/debug output mechanism visible
- Limited CLI feedback for connection errors

---

## Git Status

- **Current Branch:** MVP
- **Status:** Clean (no uncommitted changes)
- **Commits Ahead of Main:** 4
  - `f66e10d` - last human produced code
  - `4ed495d` - restrained the prompt to find important bugs
  - `c089771` - feat: add Ollama-backed AI review output with full git diff context
  - `1d76344` - uploading recent changes
  - `0fbc121` - Initial commit (main)

---

## Next Steps / Recommendations

1. **Complete MVP**
   - Implement interactive `config` command
   - Add configuration file persistence
   - Test with real repositories

2. **Quality Improvements**
   - Add comprehensive test suite
   - Improve error messages and user feedback
   - Add logging/debug modes
   - Configuration validation

3. **Feature Expansion**
   - Multi-agent review system
   - Additional provider support
   - Review output formatting options
   - Commit message generation

4. **Documentation**
   - Installation guide
   - Configuration documentation
   - Usage examples
   - Architecture documentation

---

## Summary

MOAN is a well-structured, early-stage MVP for AI-powered code review using local Ollama models. The codebase demonstrates good software engineering practices with clear separation of concerns, type safety, and thoughtful prompt engineering. The core review pipeline is functional, with Git integration working smoothly and provider abstraction allowing for future expansion. Key remaining work focuses on completing the configuration system, adding tests, and implementing planned multi-agent review features.
