# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MacStealer is a security research tool (PoC) that decrypts Chrome browser data on macOS, specifically cookies and login credentials. This is for educational and security research purposes only.

**Binary output name**: `MacStealer`
**Supported platforms**: macOS Apple Silicon only (darwin/arm64, M1 or later)
**Architecture restriction**: The tool will refuse to run on Intel Macs (amd64)

## Build System

This project uses `xgo` (github.com/crazy-max/xgo) for cross-compilation with cgo support.

**Prerequisites**: Install xgo first: `go install github.com/crazy-max/xgo@latest`

**Build commands**:
```bash
make build          # Build for darwin/arm64 only (Apple Silicon)
make test           # Run tests with race detector
make all            # Run tests and build
```

Manual build:
```bash
GOPATH=${HOME}/go/ ${HOME}/go/bin/xgo -dest bin -out MacStealer -ldflags "-s -w" -targets darwin/arm64 ./
```

## Architecture

The codebase follows a layered architecture for Chrome data extraction:

### Data Flow
1. **main.go**: CLI entry point, orchestrates the extraction flow
   - **Architecture check**: Verifies running on arm64 (Apple Silicon), exits on Intel Macs
   - Parses flags (`-kind`, `-targetpath`, `-localstate`, `-sessionstorage`)
   - Obtains Chrome master key (via macOS keychain or user-provided value)
   - Calls appropriate browsingdata extractor

2. **masterkey package**: Chrome encryption key extraction
   - `GetMasterKey()`: Invokes macOS `security` command to retrieve Chrome's keychain entry
   - `KeyGeneration()`: Derives AES key using PBKDF2 (salt: "saltysalt", iterations: 1003)
   - **Important**: This triggers a keychain access prompt on macOS

3. **browsingdata package**: SQLite database parsing and extraction
   - `GetCookie()`: Extracts and decrypts cookies from Chrome's Cookies database
   - `GetLoginData()`: Extracts and decrypts passwords from Login Data database
   - Both functions copy the database to temp location before reading (Chrome locks active DB)

4. **decrypter package**: Chrome encryption handling
   - `Chromium()`: AES-128-CBC decryption (Chrome < v80 on macOS uses this)
   - `aesGCMDecrypt()`: AES-GCM decryption (for Chrome v80+)
   - Uses hardcoded IV for CBC mode: 16 bytes of space characters (0x20)

5. **util package**: File operations and time conversion helpers

### Key Technical Details

- **Keychain integration**: Uses macOS `security find-generic-password -wa "Chrome"` command
- **Database handling**: Chrome's SQLite databases are copied before access to avoid locking issues
- **Decryption**: Chrome's encrypted values are prefixed with 3-byte header (stripped before decryption)
- **PBKDF2 parameters**: 1003 iterations, SHA1, 16-byte output (matches Chromium source)

## Running the Tool

Two operational modes:

**Mode 1: Automatic keychain access** (prompts for keychain permission):
```bash
./MacStealer-darwin-arm64 -kind cookie
./MacStealer-darwin-arm64 -kind logindata
```

**Mode 2: Manual session storage** (avoids keychain prompt):
```bash
# First, extract Chrome session storage value manually:
security find-generic-password -wa "Chrome"

# Then use it:
./MacStealer-darwin-arm64 -kind cookie -sessionstorage <value>
./MacStealer-darwin-arm64 -kind logindata -sessionstorage <value>
```

Default paths (when `-targetpath` not specified):
- Cookies: `~/Library/Application Support/Google/Chrome/Default/Cookies`
- Login Data: `~/Library/Application Support/Google/Chrome/Default/Login Data`

## Security Research Context

This is a local-only infostealer PoC that demonstrates the core techniques used by actual macOS infostealer malware. It implements the credential extraction components typically found in real infostealers, but deliberately omits the exfiltration layer (network communication, C2 connection, data upload).

**What this demonstrates:**
1. Infostealer credential harvesting techniques on macOS
2. How infostealers bypass Chrome's encryption using keychain access
3. SQLite database extraction methods used by malware
4. PBKDF2 key derivation matching Chromium's implementation

**What is intentionally excluded:**
- Network exfiltration capabilities (HTTP/HTTPS upload, C2 communication)
- Persistence mechanisms (LaunchAgents, login items)
- Anti-analysis features (VM detection, debugger checks)
- Additional data collection (browser history, autofill, crypto wallets)

This demonstrates the first stage of infostealer operation (local data harvesting) without the malicious exfiltration component. Real-world infostealers would package and transmit this data to attacker-controlled infrastructure.
