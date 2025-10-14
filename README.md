# MacStealer (Google Chrome for Apple Silicon macOS)

## NOTICE
- This software decrypt your Google Chrome's cookie, password and your creditcard data, then send them to standard output.
  - This software **does not** upload any credential to the internet.
  - This tool works local only, so not illegal model.

### Referenced source code
- This repository contains the necessary part only for PoC.

## Disclaimer
- This tool is limited to education and security research only!!

## Build & Test
- We uses `github.com/crazy-max/xgo` to build cgo binary on cross environment.
```bash
make build
```
- test
```bash
make test
```
## Supported OS and Architecture
- macOS ARM64（M1~）

## Usage
- For macOS (Normal Version)
  - (When your profile name is `Default`)
  - MacStealer asks to access keychain
    - (`security find-generic-password -wa "Chrome"` is called internally)

```bash
# Cookie
$ ./MacStealer-darwin-arm64 -kind cookie -targetpath ~/Library/Application\ Support/Google/Chrome/Default/Cookies

# Password
$ ./MacStealer-darwin-arm64 -kind logindata -targetpath ~/Library/Application\ Support/Google/Chrome/Default/Login\ Data
```

- For macOS (Use Keychain Value), When your profile name is `Default`
  1. Get `Chrome Sesssion Storage` value on Keychain
      - `security find-generic-password -wa "Chrome"`
      - or you can get the value through forensic tool like [chainbreaker](https://github.com/n0fate/chainbreaker).
  2. Decrypt cookies and passwords

```bash
# Cookie
$ ./MacStealer-darwin-arm64 -kind cookie -targetpath ~/Library/Application\ Support/Google/Chrome/Default/Cookies -sessionstorage <session storage value>

# Password
$ ./MacStealer-darwin-arm64 -kind logindata -targetpath ~/Library/Application\ Support/Google/Chrome/Default/Login\ Data -sessionstorage <session storage value>
```
