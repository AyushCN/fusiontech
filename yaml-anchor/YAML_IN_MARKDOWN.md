# Using YAML in Markdown (.md) Files

To use or display YAML inside a Markdown (`.md`) file, you generally do it in one of two ways depending on your goal. Here are the requisites for both methods:

## 1. Using YAML for Metadata (Frontmatter)
If you want to use YAML to define metadata for the Markdown file (which is heavily used by static site generators like Jekyll, Hugo, Astro, Next.js, and apps like Obsidian), you use **YAML Frontmatter**.

### Requisites:
- It **must** be the very first thing in the `.md` file.
- It **must** be surrounded by three dashes (`---`) on their own lines.

### Example:
```markdown
---
title: "My First Blog Post"
date: 2026-05-07
tags:
  - python
  - ci-cd
author: "Ayush"
draft: false
---

# Welcome to my post!
The actual markdown content starts here...
```
*(For this to actually "do" anything, the framework parsing your markdown must support frontmatter parsing, such as `gray-matter` in JavaScript or PyYAML in Python).*

---

## 2. Displaying YAML Code Snippets
If you just want to display a block of YAML code so that it looks pretty with correct syntax highlighting (like showing a GitHub Actions file to a reader), you use **Fenced Code Blocks**.

### Requisites:
- Use three backticks (```` ``` ````) to open and close the block.
- Place the word `yaml` immediately after the first set of backticks.

### Example:
````markdown
Here is an example of a GitHub Actions configuration:

```yaml
name: Python Testing
on:
  push:
    branches:
      - main
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run PyTest
        run: pytest
```
````
*(This will automatically format and color-code the snippet when rendered by GitHub, Markdown previewers, or documentation generators).*
