---
pagetitle: "Markdown Reference"
---

# Markdown Reference

This page demonstrates every Markdown feature supported by Gokesh's in-house parser. Use it to verify rendering and as a visual style reference.

---

## Headings

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

---

## Inline Formatting

Plain text, **bold text**, _italic text_, **_bold and italic_**, ~~strikethrough~~, and `inline code` all render inline.

You can combine them in a sentence: the `ParseFrontmatter` function returns a **map of strings**, _not_ a struct, so it stays ~~rigid~~ **flexible** by default.

---

## Blockquote

> Simplicity is prerequisite for reliability.
> — Edsger W. Dijkstra

> The best tool is the one you understand completely.

---

## Lists

### Unordered

- Zero external dependencies
- In-house Markdown parser
- In-house frontmatter parser
- Go standard library only
- GitHub Actions CI

### Ordered

1. Write your Markdown in `markdown/`
2. Run `make test` to build HTML into `public/`
3. Run `make dev` to preview at `localhost:8000`
4. Customize `templates/` to change the layout
5. Edit `.env` to set your author name and site title

---

## Links and Images

Visit [golang.org](https://golang.org) for Go documentation.

The Gokesh source is at [github.com/vinckr/gokesh](https://github.com/vinckr/gokesh).

![Go Gopher](https://go.dev/blog/gopher/gopher.png)

---

## Code Blocks

A fenced code block — inline formatting is **not** processed inside:

```
package main

import "fmt"

func main() {
    fmt.Println("Hello, Gokesh!")
}
```

The frontmatter parser in action:

```
---
pagetitle: "My Page"
author: vinckr
draft: false
---

Page content starts here.
```

---

## Tables

| Feature           | Implemented | Notes                        |
| ----------------- | ----------- | ---------------------------- |
| Headings          | yes         | h1 through h6                |
| Bold / Italic     | yes         | `**` and `_` syntax          |
| Strikethrough     | yes         | `~~text~~`                   |
| Inline code       | yes         | backtick syntax              |
| Links             | yes         | `[text](url)`                |
| Images            | yes         | `![alt](url)`                |
| Blockquotes       | yes         | `> text`                     |
| Unordered lists   | yes         | `-` or `*` prefix            |
| Ordered lists     | yes         | `1.` prefix                  |
| Fenced code       | yes         | triple backtick fence        |
| Tables            | yes         | pipe syntax with header row  |
| Footnotes         | no          | planned for 1.0              |
| Task lists        | no          | planned for 1.0              |

---

## Paragraphs

A paragraph is any block of text not matched by another rule. Consecutive lines are joined into a single `<p>` tag.

Blank lines separate paragraphs from each other and from other blocks.

This is the third paragraph. It contains **bold**, _italic_, and `code` inline to demonstrate that inline processing works inside paragraphs too.
