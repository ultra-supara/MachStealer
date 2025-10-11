package masterkey

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"os/exec"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrWrongSecurityCommand   = errors.New("macOS wrong security command")
	ErrCouldNotFindInKeychain = errors.New("macOS could not find in keychain")
)

func GetMasterKey(dummy string) ([]byte, error) {
	var (
		cmd            *exec.Cmd
		stdout, stderr bytes.Buffer
	)

	// Get the master key from the keychain
	// $ security find-generic-password -wa "Chrome"
	cmd = exec.Command("security", "find-generic-password", "-wa", "Chrome")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	if stderr.Len() > 0 {
		if strings.Contains(stderr.String(), "could not be found") {
			return nil, ErrCouldNotFindInKeychain
		}
		return nil, errors.New(stderr.String())
	}

	return KeyGeneration(stdout.Bytes())
}

func KeyGeneration(seed []byte) ([]byte, error) {
	chromeSecret := bytes.TrimSpace(seed)
	if chromeSecret == nil {
		return nil, ErrWrongSecurityCommand
	}
	chromeSalt := []byte("saltysalt")
	//* https://source.chromium.org/chromium/chromium/src/+/master:components/os_crypt/os_crypt_mac.mm;l=157
	key := pbkdf2.Key(chromeSecret, chromeSalt, 1003, 16, sha1.New)
	if key == nil {
		return nil, ErrWrongSecurityCommand
	}
	return key, nil
}
