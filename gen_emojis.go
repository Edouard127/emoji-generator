package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	baseAssetsUrl   = "https://raw.githubusercontent.com/twitter/twemoji/master/assets/72x72/"
	baseMappingsUrl = "https://raw.githubusercontent.com/muan/emojilib/main/dist/emoji-en-US.json"
)

type Emoji struct {
	Name string
	Png  string
}

type emojiMappings map[string][]string // key: emoji, value: [names]

func downloadMappings() (emojiMappings, error) {
	var emojiMap emojiMappings

	resp, err := http.Get(baseMappingsUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch mappings: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&emojiMap)
	if err != nil {
		return nil, err
	}

	return emojiMap, nil
}

func constructEmojis(emojiMap emojiMappings) ([]Emoji, error) {
	var emojis []Emoji

	for emoji, names := range emojiMap {
		emojis = append(emojis, Emoji{
			Name: toSnakeCase(names[0]),
			Png:  baseAssetsUrl + emojiToHex(emoji) + ".png",
		})
	}

	return emojis, nil
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s, the mappings are most likely outdated", resp.Status)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// emojiToHex converts an emoji to a hex string.
// It only returns the first part of the emoji if it has multiple parts.
// This way, we can use the same emoji for different skin tones.
func emojiToHex(emoji string) (output string) {
	var hexParts []string
	for _, r := range []rune(emoji) {
		if r == 0xfe0f {
			// Skip the variation selector
			continue
		}
		hexParts = append(hexParts, fmt.Sprintf("%04x", r))
	}

	output = strings.Join(hexParts, "-")
	return
}

var reg = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func toSnakeCase(input string) string {
	return strings.ToLower(reg.ReplaceAllString(input, "_"))
}
