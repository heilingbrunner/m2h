# m2h

A small, simple markdown to htm/html converter.

## Features

- converts a markdown file (or stdin) to a single HTML document on stdout
- converts any source file as a highlighted code page with `-l <language>`
- embeds local images as base64 data URIs
- GitHub Flavored Markdown (tables, strikethrough, task lists, autolinks) with auto heading IDs
- syntax highlighting via chroma, selectable style with `-s` (e.g. `-s dracula`)
- auto-detects `mermaid` and `d2` diagrams and wires up the required scripts
- KaTeX math rendering (auto-detected)
- optional paged.js layout for nicer PDF printing with `-p` (not compatible with mermaid)
- open the result straight in the browser with `-b`

## Usage

```bash
$ m2h file1.md > file1.htm
```
