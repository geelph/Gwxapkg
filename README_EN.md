# Gwxapkg

<div align="center">

![Version](https://img.shields.io/badge/version-2.7.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)
![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)

**[中文](README.md) | [English](README_EN.md) | [日本語](README_JA.md)**

A Go-based WeChat Mini Program `.wxapkg` unpacker with automatic scanning, decryption, decompilation, and security analysis.

</div>

---

## Important Disclaimer

> **This tool is intended only for legally authorized security research, reverse engineering, compatibility verification, learning, and auditing of assets you own or are explicitly authorized to assess.**
>
> **Do not use this tool for any unauthorized scenario, including but not limited to unpacking third-party Mini Programs without permission, code or asset theft, bulk scraping, privacy extraction, bypassing platform protections, offensive testing, commercial abuse, risk-control evasion, malware delivery, or any activity that violates laws, platform terms, copyright rules, privacy policies, or internal compliance requirements.**
>
> **Before using this tool, you are responsible for confirming that you have clear and valid authorization for the target Mini Program, account, device, environment, business system, and data involved, and for evaluating any legal, compliance, copyright, privacy, data-leakage, account-ban, service-interruption, or third-party claim risks.**
>
> **The authors and contributors of this project are not liable for any direct or indirect loss arising from the use, misuse, or abuse of this tool, including but not limited to data leakage, privacy infringement, intellectual property disputes, platform penalties, account suspension, system failures, production incidents, financial loss, administrative penalties, or criminal liability.**
>
> **If you cannot verify that your use is authorized, stop immediately. Continuing to use this tool means you understand and accept the above risks and responsibilities.**

---

## Key Features

### Smart Unpacking
- Auto-scan WeChat Mini Program cache directories on macOS and Windows
- Auto-decrypt encrypted `wxapkg` files from desktop WeChat cache
- One-command processing for a specified AppID
- Proper handling of main packages and subpackages

### Code Restoration
- Full restoration for `wxml`, `wxss`, `js`, `json`, and `wxs`
- Automatic JavaScript/CSS/HTML formatting
- Default JavaScript deobfuscation for common string-array, `\xNN`, `\uNNNN`, and hexadecimal literal patterns
- Restore the original Mini Program project structure
- Extract images, audio, video, and other resources

### Security Analysis
- 200+ built-in sensitive-data detection rules
- Better false-positive filtering and deduplication
- URL / API endpoint extraction with optional Postman Collection export
- Excel and HTML reports with file paths and line numbers
- Obfuscated-file reporting with restoration status

### Performance
- Dynamic worker count based on CPU cores
- Buffered file output
- Regex precompilation
- Optimized release builds

---

## Supported File Types

| File Type | Support | Description |
|-----------|---------|-------------|
| `.wxml` | Yes | Page structure restoration |
| `.wxss` | Yes | Style restoration |
| `.js` | Yes | Restoration, formatting, and default deobfuscation |
| `.json` | Yes | Config extraction |
| `.wxs` | Yes | WXS restoration |
| Images / Audio / Video | Yes | Resource extraction |

---

## Installation

### Download Prebuilt Binary

Download the appropriate binary from [Releases](https://github.com/25smoking/Gwxapkg/releases).

### Build from Source

```bash
git clone https://github.com/25smoking/Gwxapkg.git
cd Gwxapkg

go build -ldflags="-s -w" -o gwxapkg .
```

Requirement: Go 1.21 or later.

---

## Quick Start

```bash
# Scan cached Mini Programs
./gwxapkg scan

# Scan with cache-path diagnostics
./gwxapkg scan --verbose

# Process a specific AppID
./gwxapkg all -id=<AppID>

# Process a specific AppID and export Postman Collection
./gwxapkg all -id=<AppID> -postman

# Scan an already-unpacked directory only
./gwxapkg scan-only -dir=./output/<AppID> -format=both -postman

# Unpack a single wxapkg file
./gwxapkg -id=<AppID> -in=<file_path>

# Repack
./gwxapkg repack -in=<directory_path>
```

### Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-id` | Mini Program AppID | - |
| `-in` | Input file or directory | - |
| `-out` | Output directory | auto |
| `-restore` | Restore project directory structure | true |
| `-pretty` | Beautify output | true |
| `-sensitive` | Generate Excel / HTML sensitive reports | true |
| `-postman` | Export `api_collection.postman_collection.json` | false |
| `-workspace` | Keep internal workspace for precise repack | false |
| `--verbose` | Show scan diagnostics for cache candidates | false |

---

## Default Output Directory

If `-out` is not provided, results are written to `output/<AppID>`.

- Compiled binary: under the executable directory
- `go run .`: under the current working directory
- Interactive `scan`: also uses `output/<AppID>`

Example:

```text
/Applications/Gwxapkg/output/wx1234567890abcdef
./output/wx1234567890abcdef
```

Typical output:

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

## Security Reports

### Report Outputs

- `-sensitive=true` generates:
  - `sensitive_report.xlsx`
  - `sensitive_report.html`
- `-postman=true` generates:
  - `api_collection.postman_collection.json`
- `-postman` is independent from `-sensitive`
- `scan-only` reuses the same scanner and deobfuscation pipeline

### Obfuscated File Reporting

When JavaScript matches supported obfuscation patterns, the report records:

- File path
- Score
- Techniques
- Status: `restored`, `partial`, or `flagged`
- Output tag: `[OBFUSCATED] ...`

### API Extraction

- URLs and API endpoints are extracted from the same scan report
- Relative paths remain unchanged
- If the HTTP method cannot be inferred safely, it is exported as `UNKNOWN`

---

## Rule Configuration

- Built-in rules are enabled by default
- The tool no longer auto-generates `config/rule.yaml`
- If you manually provide `config/rule.yaml`, your custom rules override the built-in set

---

## v2.7.0 Highlights

- Postman Collection export
- Default JavaScript deobfuscation
- Obfuscated-file reporting in Excel / HTML
- `scan-only` mode
- `scan --verbose` and `all --verbose`
- Safer unpacking with path validation, hard limits, and staged errors
- Built-in rules enabled by default
- Manual GitHub Actions triggers for CI and Release

For the full release summary, see [RELEASE_NOTES.md](RELEASE_NOTES.md).

---

## GitHub Actions

This repository includes two workflows:

- `CI`
  - Triggered by push to `main`, pull requests to `main`, or manual run
  - Runs `go build` and `go test`
- `Release`
  - Triggered by version tags like `v2.7.0` or manual run
  - Builds Windows, Linux, macOS Intel, and macOS Apple Silicon binaries
  - Publishes them to GitHub Releases

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
