# MOAN

MOAN – Modular Orchestration for Automated Nodes

MOAN is a local-first AI code review CLI. It analyzes your Git changes, builds a structured review prompt, and sends the diff to an AI model to find bugs, security risks, reliability problems, and maintainability issues before you commit.

MOAN works with local models through Ollama, and with OpenAI and Anthropic APIs.

---

## Requirements

- Go 1.26+ (to build)
- Git
- One of:
  - Ollama running locally (default)
  - An OpenAI API key
  - An Anthropic API key

Recommended local model:

```bash
ollama pull qwen2.5-coder:7b
```

---

## Install / Build

```bash
git clone https://github.com/SPCDIAZRIVERACHRISTIAN/moan.git
cd moan
go build -o moan .
```

Optionally move the binary onto your PATH:

```bash
sudo mv moan /usr/local/bin/
```

---

## Quick start

```bash
# Create a default config (ollama, qwen2.5-coder:7b)
moan config init

# Check everything is wired up
moan validate

# Review your current Git changes
moan review
```

---

## Configuration

The config file lives at:

- Linux/macOS: `~/.config/moan/config.yaml`
- Windows: `%AppData%\moan\config.yaml`

Override the location with `MOAN_CONFIG=/path/to/config.yaml`.

Example config:

```yaml
provider: ollama
model: ""               # optional global override; usually leave empty
system_prompt: ""
timeout_seconds: 120
max_diff_chars: 60000
max_files: 30

ollama:
  base_url: "http://localhost:11434"
  model: "qwen2.5-coder:7b"

openai:
  api_key: ""
  base_url: "https://api.openai.com/v1"
  model: "gpt-4o-mini"

anthropic:
  api_key: ""
  base_url: "https://api.anthropic.com"
  model: "claude-3-5-haiku-latest"
```

### Config commands

```bash
moan config init                      # create default config file
moan config show                      # show effective config (keys masked)
moan config setup                     # interactive setup
moan config set-provider ollama       # or openai / anthropic
moan config set-model <model>         # sets the model for the current provider
moan config set-openai-key <key>
moan config set-anthropic-key <key>
moan config set-ollama-url <url>
moan config set-timeout <seconds>
```

### Environment variables

Environment variables override the config file (and CLI flags override both):

| Variable | Purpose |
|---|---|
| `MOAN_CONFIG` | Config file path |
| `MOAN_PROVIDER` | Provider (`ollama`, `openai`, `anthropic`) |
| `MOAN_MODEL` | Model override |
| `OLLAMA_BASE_URL` | Ollama base URL |
| `OPENAI_API_KEY` | OpenAI key |
| `OPENAI_BASE_URL` | OpenAI base URL |
| `ANTHROPIC_API_KEY` | Anthropic key |
| `ANTHROPIC_BASE_URL` | Anthropic base URL |

---

## Usage

### Ollama (default)

```bash
ollama serve
moan review
```

### OpenAI

```bash
moan config set-openai-key sk-...
moan review --provider openai --model gpt-4o-mini
```

### Anthropic / Claude

```bash
moan config set-anthropic-key sk-ant-...
moan review --provider anthropic --model claude-3-5-haiku-latest
```

`--provider` and `--model` affect that run only; they never modify the config file. Use `moan config set-provider` to change the default.

### Review flags

```bash
moan review --staged              # review staged changes only
moan review --output markdown     # markdown output (default: text)
moan review --timeout 180         # request timeout in seconds
moan review --max-diff-chars 80000
moan review --max-files 50
moan review --allow-truncate      # truncate oversized diffs instead of failing
moan review --dry-run             # build the request, show sizes, don't call the model
moan review --debug               # internal info (API keys are always masked)
moan review --save                # save the review to .moan/reviews/<timestamp>.md
```

### Validation

```bash
moan validate
```

```text
MOAN VALIDATION
---------------
Git repository: OK
Git changes: OK (3 files)
Config: OK (/home/you/.config/moan/config.yaml)
Provider: openai
Model: gpt-4o-mini
API key: OK
Connection: OK

STATUS: READY
```

---

## Example review output

```text
STATUS: READY
Provider: ollama
Model: qwen2.5-coder:7b
Review scope: staged + unstaged
Files changed: 2

- internal/foo/bar.go | +14 -3
- cmd/review.go | +2 -1

AI REVIEW
-----------------
STATUS: NEEDS_CHANGES

SUMMARY:
The diff adds a retry loop to bar.go and a flag to the review command.

FINDINGS:

### Finding 1
File: internal/foo/bar.go
Severity: high
Category: bug
Issue: The retry loop never increments the attempt counter.
Evidence: `for attempts < 3 { ... }` with no `attempts++` inside the loop.
Impact: Infinite loop on persistent failure.
Suggested fix: Increment `attempts` at the end of each iteration.

FINAL DECISION:
NEEDS_CHANGES due to the infinite retry loop.

SUMMARY
-------
Status: NEEDS WORK
Critical: 0
High: 1
Medium: 0
Low: 0
```

---

## Exit codes

MOAN is scriptable:

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Setup/runtime error |
| 2 | Validation failed |
| 3 | No Git changes |
| 4 | Provider/API error |

---

## Security notes

- API keys are stored only in your local config file (`0600` permissions) when you explicitly set them.
- `moan config show` and `--debug` always mask keys.
- Only the Git diff (plus changed file names/stats) is sent to the provider — never the full repository.

---

## MVP limitations

- Untracked (new, never-added) files are not included in the diff; `git add` them first.
- One review pass with a single model — no multi-agent review yet.
- Output formats are text and markdown only.
- Diff truncation is character-based, not file-aware.
- No commit message generation, no HTML reports, no TUI.
