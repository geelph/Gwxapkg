# Gwxapkg

<div align="center">

![Version](https://img.shields.io/badge/version-2.7.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)

**[中文](README.md) | [English](README_EN.md) | [日本語](README_JA.md)**

Go で実装された WeChat ミニプログラム `.wxapkg` 解析ツールです。自動スキャン、復号、逆コンパイル、セキュリティ分析に対応しています。

</div>

---

## 重要な免責事項

> **本ツールは、正当な権限のあるセキュリティ調査、リバースエンジニアリング、互換性検証、学習、および自社資産または明示的に許可を受けた対象の監査にのみ使用してください。**
>
> **許可のないミニプログラムの解凍、コードやアセットの盗用、大量収集、プライバシー情報の抽出、プラットフォーム保護の回避、攻撃的なテスト、商用悪用、リスク制御の回避、マルウェア拡散、ならびに法令、プラットフォーム規約、著作権ルール、プライバシーポリシー、社内コンプライアンス要件に反する行為には使用しないでください。**
>
> **利用者は、対象となるミニプログラム、アカウント、端末、環境、業務システム、および関連データについて、明確かつ有効な権限を有していることを自ら確認し、法的リスク、コンプライアンスリスク、著作権リスク、プライバシーリスク、情報漏えい、アカウント停止、サービス停止、第三者からの請求などの影響を自ら評価する責任があります。**
>
> **本プロジェクトの作者およびコントリビューターは、本ツールの使用、誤用、または乱用によって生じるいかなる直接的または間接的損害についても責任を負いません。これには、情報漏えい、プライバシー侵害、知的財産紛争、プラットフォームからの制裁、アカウント停止、システム障害、本番事故、経済的損失、行政処分、刑事責任が含まれますが、これらに限りません。**
>
> **利用が正当な権限に基づくか確認できない場合は、直ちに使用を中止してください。使用を継続した場合、上記のリスクと責任を理解し、受け入れたものとみなされます。**

---

## 主な機能

### スマート解凍
- macOS / Windows の WeChat ミニプログラムキャッシュを自動スキャン
- デスクトップ版 WeChat の暗号化 `wxapkg` を自動復号
- AppID 指定で関連パッケージをまとめて処理
- メインパッケージとサブパッケージを適切に処理

### コード復元
- `wxml`、`wxss`、`js`、`json`、`wxs` を復元
- JavaScript / CSS / HTML を自動整形
- 代表的な文字列配列、`\xNN`、`\uNNNN`、16 進数リテラルに対する JavaScript の既定反混淆
- 元のミニプログラム構成に近いディレクトリを復元
- 画像、音声、動画などのリソースを抽出

### セキュリティ分析
- 200 以上の内蔵検出ルール
- 誤検知低減と重複除去
- URL / API エンドポイント抽出と Postman Collection 出力
- ファイルパスと行番号付きの Excel / HTML レポート
- 混淆ファイルの復元状態レポート

### パフォーマンス
- CPU コア数に応じた動的並列処理
- バッファ付きファイル出力
- 正規表現の事前コンパイル
- 最適化されたリリースビルド

---

## 対応ファイル形式

| ファイル形式 | 対応 | 説明 |
|-------------|------|------|
| `.wxml` | はい | ページ構造の復元 |
| `.wxss` | はい | スタイルの復元 |
| `.js` | はい | 復元、整形、既定反混淆 |
| `.json` | はい | 設定抽出 |
| `.wxs` | はい | WXS 復元 |
| 画像 / 音声 / 動画 | はい | リソース抽出 |

---

## インストール

### ビルド済みバイナリを利用する

[Releases](https://github.com/25smoking/Gwxapkg/releases) から対象プラットフォーム向けバイナリを取得してください。

### ソースからビルドする

```bash
git clone https://github.com/25smoking/Gwxapkg.git
cd Gwxapkg

go build -ldflags="-s -w" -o gwxapkg .
```

必要環境: Go 1.21 以上

---

## クイックスタート

```bash
# キャッシュされたミニプログラムを一覧表示
./gwxapkg scan

# キャッシュ候補パスの診断を表示
./gwxapkg scan --verbose

# AppID を指定して処理
./gwxapkg all -id=<AppID>

# AppID を指定し、Postman Collection も出力
./gwxapkg all -id=<AppID> -postman

# 既に解凍済みのディレクトリだけを再スキャン
./gwxapkg scan-only -dir=./output/<AppID> -format=both -postman

# 単一の wxapkg を解凍
./gwxapkg -id=<AppID> -in=<file_path>

# 再パック
./gwxapkg repack -in=<directory_path>
```

### よく使うオプション

| オプション | 説明 | 既定値 |
|------------|------|--------|
| `-id` | ミニプログラム AppID | - |
| `-in` | 入力ファイルまたはディレクトリ | - |
| `-out` | 出力先ディレクトリ | 自動 |
| `-restore` | プロジェクト構造を復元 | true |
| `-pretty` | 出力を整形 | true |
| `-sensitive` | Excel / HTML の機密情報レポートを生成 | true |
| `-postman` | `api_collection.postman_collection.json` を生成 | false |
| `-workspace` | 正確な再パック用ワークスペースを保持 | false |
| `--verbose` | スキャン候補パスの診断を表示 | false |

---

## 既定の出力ディレクトリ

`-out` を指定しない場合、出力先は `output/<AppID>` になります。

- コンパイル済みバイナリ実行時: 実行ファイルのある場所の下
- `go run .` 実行時: 現在の作業ディレクトリの下
- 対話式 `scan` でも同じ規則を使用

例:

```text
/Applications/Gwxapkg/output/wx1234567890abcdef
./output/wx1234567890abcdef
```

典型的な出力:

```text
output/
└── wx1234567890abcdef/
    ├── app.js
    ├── page-frame.html
    ├── sensitive_report.xlsx
    ├── sensitive_report.html
    ├── api_collection.postman_collection.json
    └── .gwxapkg/
```

---

## セキュリティレポート

### 出力されるレポート

- `-sensitive=true` で生成:
  - `sensitive_report.xlsx`
  - `sensitive_report.html`
- `-postman=true` で生成:
  - `api_collection.postman_collection.json`
- `-postman` は `-sensitive` と独立して利用可能
- `scan-only` でも同じスキャナと反混淆処理を再利用

### 混淆ファイルの報告

JavaScript が対応パターンに一致した場合、レポートには次が記録されます。

- ファイルパス
- スコア
- 技術パターン
- 状態: `restored` / `partial` / `flagged`
- 出力タグ: `[OBFUSCATED] ...`

### API 抽出

- URL と API エンドポイントは同じスキャン結果から抽出
- 相対パスはそのまま保持
- HTTP メソッドを安全に推定できない場合は `UNKNOWN` として出力

---

## ルール設定

- 内蔵ルールは既定で有効です
- `config/rule.yaml` は自動生成されません
- 手動で `config/rule.yaml` を配置した場合、その内容が内蔵ルールより優先されます

---

## v2.7.0 の主な強化点

- Postman Collection 出力
- JavaScript の既定反混淆
- Excel / HTML での混淆ファイル報告
- `scan-only` モード
- `scan --verbose` と `all --verbose`
- パス検証、ハード上限、段階別エラーによる安全な解凍
- 内蔵ルールの既定有効化
- GitHub Actions の手動 CI / Release 起動

詳細は [RELEASE_NOTES.md](RELEASE_NOTES.md) を参照してください。

---

## GitHub Actions

このリポジトリには 2 つのワークフローがあります。

- `CI`
  - `main` への push、`main` 向け pull request、または手動実行で起動
  - `go build` と `go test` を実行
- `Release`
  - `v2.7.0` のようなタグ push、または手動実行で起動
  - Windows、Linux、macOS Intel、macOS Apple Silicon 向けバイナリを生成
  - GitHub Releases に公開

---

## ライセンス

本プロジェクトは MIT License で提供されています。詳細は [LICENSE](LICENSE) を参照してください。
