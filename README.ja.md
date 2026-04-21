# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[English](README.md) | 日本語

[![Demo](https://i.gyazo.com/a5f8265a0865e5a37dc83733ca61069a.gif)](https://gyazo.com/a5f8265a0865e5a37dc83733ca61069a)

> Go 製の超高速 Markdown リンター — **100,000 行以上を約 170ms** で処理。シングルバイナリで Node.js 不要、HTTP リンクバリデーション機能を内蔵。

**かんたんインストール**（macOS / Linux）:

```sh
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh
```

**バイナリをダウンロード**（Go 環境不要）:

[GitHub Releases](https://github.com/shinagawa-web/gomarklint/releases/latest) からお使いのプラットフォーム向けバイナリをダウンロードできます。

```sh
# macOS / Linux
tar -xzf gomarklint_Darwin_x86_64.tar.gz
sudo mv gomarklint /usr/local/bin/
# sudo が使えない場合はユーザーローカルへ
mkdir -p ~/.local/bin && mv gomarklint ~/.local/bin/
```

```powershell
# Windows (PowerShell)
Expand-Archive -Path gomarklint_Windows_x86_64.zip -DestinationPath "$env:LOCALAPPDATA\Programs\gomarklint"
# PATH に追加（初回のみ）
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:LOCALAPPDATA\Programs\gomarklint", "User")
```

**Homebrew を使う場合:**

```sh
brew install shinagawa-web/tap/gomarklint
```

**npm を使う場合:**

```sh
npm install -g @shinagawa-web/gomarklint
```

**`go install` を使う場合:**

```sh
go install github.com/shinagawa-web/gomarklint/v2@latest
```

- **100,000 行以上を約 170ms** で処理 — JIT ウォームアップなし、ランタイムオーバーヘッドなし。
- リンク切れや見出しの問題をドキュメント公開前にキャッチ。
- 予測可能な構造を強制（「なぜ H2 の下に H4 があるの？」をなくす）。
- 人間にも機械にも優しい出力（JSON 対応）。

## CI 連携

### GitHub Actions

```yaml
name: gomarklint

on:
  push:
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: shinagawa-web/gomarklint-action@v1
        with:
          args: '.'
```

オプションの詳細: [gomarklint-action](https://github.com/shinagawa-web/gomarklint-action)

### pre-commit

`.pre-commit-config.yaml` に追加:

```yaml
repos:
  - repo: https://github.com/shinagawa-web/gomarklint
    rev: v2.8.0
    hooks:
      - id: gomarklint
```

## ドキュメント

完全なドキュメントは **[https://shinagawa-web.github.io/gomarklint/](https://shinagawa-web.github.io/gomarklint/)** で参照できます。

- [クイックスタート](https://shinagawa-web.github.io/gomarklint/docs/quick-start/)
- [ルール一覧](https://shinagawa-web.github.io/gomarklint/docs/rules/)
- [CLI リファレンス](https://shinagawa-web.github.io/gomarklint/docs/cli/)
- [設定](https://shinagawa-web.github.io/gomarklint/docs/configuration/)
- [GitHub Actions 連携](https://shinagawa-web.github.io/gomarklint/docs/github-actions/)

## コントリビュート

Issue・提案・PR 歓迎です！

必要環境: Go `1.22+`（最新の安定版を推奨）

```sh
make test      # ユニットテスト
make test-e2e  # エンドツーエンドテスト
make build     # バイナリビルド
```

### Git フック

プッシュ前に自動でlintとユニットテストを実行するpre-pushフックをインストールできます：

```sh
make install-hooks
```

緊急時にフックをスキップする場合：

```sh
git push --no-verify
```

## ライセンス

MIT License
