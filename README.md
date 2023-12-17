# git-gpt

## Introduction

Harness the power of natural language generation with this innovative Git plugin that seamlessly integrates with ChatGPT to automate the creation of commit messages and code reviews. Say goodbye to the hassle of crafting detailed messages manually—let the language model do the heavy lifting for you.

## Features

🤖 Automated Commit Messages: Generate informative and context-aware commit messages with just a simple command. Whether it's bug fixes, feature additions, or documentation updates, ChatGPT has got your back.

👥 Intelligent Code Reviews: Elevate your code review process by utilizing ChatGPT to provide insightful and constructive feedback. Enhance collaboration among your team members with AI-assisted code critiques.

🔍 Contextual Understanding: Benefit from ChatGPT's contextual understanding to ensure that generated messages align with your project's unique coding conventions and style guidelines.

🌐 Multi-Lingual Support: Collaborate seamlessly across language barriers with ChatGPT's multilingual capabilities. Generate commit messages and reviews in the language of your choice.

## Getting Started

### Installation

**Using Go:**

Install using go command:

```bash
go install github.com/rammstein4o/git-gpt@latest
```

**Using binaries:**

Download the binary appropriate for your platform. Rename it to `git-gpt`, put it in your PATH and make it executable:

* Linux: [git-gpt-linux-amd64](dist/git-gpt-linux-amd64)
* MacOS: [git-gpt-darwin-amd64](dist/git-gpt-darwin-amd64)
* Windows: [git-gpt-win-amd64](dist/git-gpt-win-amd64)

### Configuration

TODO

### Usage

Modify some files in your repo. Then run:

```bash
git add -A
git gpt commit
```

## License

MIT