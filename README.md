# gomarklint

> A fast and lightweight Markdown linter written in Go.

**gomarklint** checks your Markdown files for common issues such as heading structure problems, trailing blank lines, unclosed code blocks, and more. Designed to be minimal, fast, and CI-friendly.

---

## ✨ Features

- ✅ Lint individual `.md` files or entire directories
- ✅ Checks for heading level consistency (`# → ## → ###`)
- ✅ Detects missing trailing blank lines
- ✅ Detects unclosed code blocks
- ⚡️ Blazing fast — built with Go

📝 **Note:** By default, `gomarklint` assumes heading levels start from `##` (H2), not `#` (H1), to align with common blog and static site conventions.
---

## 📦 Installation (for local testing)

```bash
go install github.com/yourname/gomarklint@latest
```

Or clone and run:

```bash
git clone https://github.com/yourname/gomarklint.git
cd gomarklint
go run main.go ./README.md
```

## 🚀 Usage

```bash
gomarklint ./posts --min-heading 2
gomarklint ./posts ./docs
gomarklint ./content --ignore CHANGELOG.md --json
```

Options:

- `--min-heading` — Set the minimum heading level to expect. Defaults to `2` (i.e. `##`), which aligns with common blogging/static site practices.

## 🛣 Roadmap

v0.1.0
- [x] Basic CLI setup with Cobra
- [x] Lint single file
- [x] Check heading level jumps
- [ ] Check missing final blank lines
- [ ] Check unclosed code blocks

v0.2.0
- [ ] Support multiple files and directories
- [ ] Output file name and line number
- [ ] Recursively search .md files

v0.3.0
- [ ] Add rules: duplicate headings, empty alt text, TODO comments
- [ ] Add --ignore flag
- [ ] Add --json output option

v0.4.0
- [ ] GitHub Actions support
- [ ] Cross-platform binaries via goreleaser

v1.0.0
- [ ] At least 5 rules with test coverage
- [ ] Stable CLI interface
- [ ] Prebuilt binaries for macOS/Linux/Windows
- [ ] Clear README and blog post with real usage examples

## 📁 Project Structure

```
gomarklint/
├── cmd/             # CLI entrypoint (Cobra)
├── internal/
│   ├── rule/        # Individual lint rules
│   └── parser/      # Markdown parsing logic
├── testdata/        # Sample Markdown files
├── main.go
└── README.md
```

## 🤝 Contributing

Issues, suggestions, and PRs are welcome!
This project is just getting started and will evolve step by step.

## 📜 License

MIT License
