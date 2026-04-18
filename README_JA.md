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

## ⚠️ 重要な法的および利用上の注意

**本ツールは、合法かつ正当で、かつ十分な権限を取得している場合に限り、セキュリティ調査、リバースエンジニアリング、互換性検証、学習交流、および自己所有または受託資産の技術監査のために使用することができます。**

**本ツールを使用する前に、すべての利用者は、対象となるミニプログラム、関連アカウント、端末、データ、ネットワーク環境、業務システム、その他の関連資産について、明確かつ継続的で、証明可能な法的権限を有していることを自ら確認しなければなりません。権限の範囲または有効性を確認できない場合は、直ちに使用を中止してください。**

### 禁止される用途

**本ツールは、無権限の利用や、法令、プラットフォーム規約、契約上の義務、知的財産権、またはプライバシー保護に違反するおそれのあるいかなる場面でも使用してはなりません。これには以下が含まれますが、これらに限られません。**

- **許可なく第三者のミニプログラムのコード、アセット、またはデータを解凍、解析、抽出、複製、配布すること**
- **プラットフォーム保護、リスク制御措置、アクセス制限、その他の技術的制御を回避すること**
- **大量収集、データスクレイピング、プライバシー情報の抽出、API の乱用、または自動化された攻撃的テスト**
- **商業的盗用、悪意ある模倣、悪意ある配布、不正産業での利用、その他の不適切な目的**
- **他者の権利侵害、プラットフォームからの制裁、業務中断、情報漏えい、またはコンプライアンスリスクを引き起こし得る行為**

### リスクに関する注意

**本ツールの利用には、法的責任、行政処分、民事賠償、刑事リスク、知的財産紛争、プライバシー侵害、アカウント停止、サービス中断、情報漏えい、本番事故、第三者からの請求などのリスクが伴う可能性があります。**

**利用者は、本ツールの使用、誤用、または乱用に起因する一切のリスクおよび結果を自ら評価し、自ら負担するものとします。**

### 責任の制限

**適用法令で認められる最大限の範囲において、本プロジェクトならびにその作者、コントリビューター、メンテナーは、本ツールの使用不能、使用、誤用、または乱用に起因するいかなる直接的、間接的、付随的、特別、懲罰的、または結果的損害についても責任を負いません。これには、データ損失、情報漏えい、プライバシー侵害、知的財産紛争、プラットフォーム制裁、アカウント停止、システム障害、業務損失、経済的損失、行政責任、刑事責任が含まれますが、これらに限られません。**

**本ツールは「現状有姿」で提供され、商品性、特定目的適合性、安定性、正確性、完全性、継続的利用可能性、適法性を含む一切の明示または黙示の保証を伴いません。**

### 使用継続は同意を意味します

**本ツールを引き続きダウンロード、インストール、実行、配布、または使用することにより、利用者は本通知の全文を読み、理解し、同意したものとみなされ、かつ合法的に権限を有する範囲内でのみ本ツールを使用することを約束するものとします。**

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
