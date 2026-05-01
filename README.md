# MOAN

MOAN – Modular Orchestration for Automated Nodes 

MOAN is a local-first code review CLI that analyzes your Git changes and sends them to an AI model for review.

The goal is to create a developer tool that can inspect local diffs, identify bugs, architecture problems, maintainability issues, and security risks before code is committed.

MOAN is designed to work with local AI models through Ollama first, with future support for multiple providers and specialized review agents.

---

## Current Status

MOAN is in early MVP development.

Current capabilities:

- Validate that the current directory is a Git repository
- Detect whether there are tracked file changes
- Read changed file metadata from Git
- Connect to a local Ollama model
- Send review requests to a configured model
- Return AI-generated review feedback

Planned capabilities:

- Full Git diff review
- Multi-agent review system
- Security-focused reviewer
- Architecture-focused reviewer
- Test/QA reviewer
- Interactive config command
- Review reports
- Commit message generation
- Human approval before pushing changes

---

## Requirements

You need:

- Go installed
- Git installed
- Ollama installed and running
- A local or remote Ollama model available

Recommended model for code review:

```bash
ollama pull qwen2.5-coder:7b
