![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/ngyewch/mdbook-asciidoc/build.yml)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/ngyewch/mdbook-asciidoc)
![GitHub last commit](https://img.shields.io/github/last-commit/ngyewch/mdbook-asciidoc)

# mdbook-asciidoc

A backed for [mdbook](https://rust-lang.github.io/mdBook/) that outputs [AsciiDoc](https://asciidoc.org/)

## Installation

To use, install the tool.

### Via [`ubi`](https://github.com/houseabsolute/ubi)

```
ubi -p ngyewch/mdbook-asciidoc
```

### Via `go`

```
go install github.com/ngyewch/mdbook-asciidoc@latest
```

## Setup

Next you need to let `mdbook` know to use the alternate renderer by updating your `book.toml` file. This is done by simply adding an empty `output.asciidoc` table.

```
[output.asciidoc]
```

## Configuration

| Name                | Type  | Default | Description                           |
|---------------------|-------|---------|---------------------------------------|
| `min-heading-level` | `int` |         | Minimum heading level to be rendered. |
