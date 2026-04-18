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

## ⚠️ Legal and Usage Notice

**This tool may only be used for security research, reverse engineering, compatibility verification, learning, and technical auditing of self-owned or entrusted assets, and only where such use is lawful, proper, and fully authorized.**

**Before using this tool, every user must independently confirm that they have clear, continuing, and demonstrable legal authorization over the target Mini Program, related accounts, devices, data, network environment, business systems, and any other associated resources. If you cannot confirm the scope or validity of such authorization, you must stop using this tool immediately.**

### Prohibited Uses

**This tool must not be used in any unauthorized scenario or in any way that may violate applicable laws, platform rules, contractual obligations, intellectual property rights, or privacy protections, including but not limited to:**

- **Unpacking, analyzing, extracting, copying, or distributing third-party Mini Program code, assets, or data without permission**
- **Bypassing platform protections, risk-control mechanisms, access restrictions, or any other technical controls**
- **Bulk collection, data scraping, privacy extraction, API abuse, or automated offensive testing**
- **Commercial theft, malicious imitation, malicious distribution, abuse in illicit industries, or any other improper purpose**
- **Any conduct that may cause harm to others, platform penalties, business interruption, data leakage, or compliance risks**

### Risk Notice

**Use of this tool may involve and result in risks including but not limited to legal liability, administrative penalties, civil damages, criminal exposure, intellectual property disputes, privacy infringement claims, account suspension, service interruption, data leakage, production incidents, and third-party claims.**

**Users are solely responsible for evaluating and bearing all risks and consequences arising from their use, misuse, or abuse of this tool.**

### Limitation of Liability

**To the maximum extent permitted by applicable law, this project and its authors, contributors, and maintainers shall not be liable for any direct, indirect, incidental, special, punitive, or consequential loss arising from the use, inability to use, misuse, or abuse of this tool, including but not limited to data loss, information leakage, privacy infringement, intellectual property disputes, platform penalties, account suspension, system failures, business loss, financial loss, administrative liability, or criminal liability.**

**This tool is provided on an “as is” basis, without any express or implied warranty, including but not limited to warranties of merchantability, fitness for a particular purpose, stability, accuracy, completeness, continued availability, or legal suitability.**

### Use Constitutes Acceptance

**By continuing to download, install, run, distribute, or use this tool, you acknowledge that you have read, understood, and agreed to this notice in full, and that you will use this tool only within a lawfully authorized scope.**

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
