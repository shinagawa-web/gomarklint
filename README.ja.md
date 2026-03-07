# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[English](README.md) | 日本語

> エンジニアリングチームのための高速・実用的な Markdown リンター。Go 製、CI 向け設計。

```sh
go install github.com/shinagawa-web/gomarklint@latest
```

- リンク切れや見出しの問題をドキュメント公開前にキャッチ。
- 予測可能な構造を強制（「なぜ H2 の下に H4 があるの？」をなくす）。
- 人間にも機械にも優しい出力（JSON 対応）。
- **100,000 行以上を約 170ms** で処理 — ローカル開発にもCI にも十分な速さ。

## ドキュメント

完全なドキュメントは **[https://shinagawa-web.github.io/gomarklint/](https://shinagawa-web.github.io/gomarklint/)** で参照できます。

- [クイックスタート](https://shinagawa-web.github.io/gomarklint/docs/)
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

## ライセンス

MIT License
