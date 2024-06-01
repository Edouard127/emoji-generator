package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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
		hexName := emojiToHex(emoji)

		emojis = append(emojis, Emoji{
			Name: names[0], // Take the first name
			Png:  baseAssetsUrl + hexName + ".png",
		})
	}

	return emojis, nil
}

func main() {
	fmt.Println("Downloading mappings...")
	emojiMap, err := downloadMappings()
	if err != nil {
		panic(err)
	}

	fmt.Println("Downloading emojis...")
	emojis, err := constructEmojis(emojiMap)
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing emojis to disk...")
	err = os.MkdirAll("emojis", 0755)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(emojis))

	start := time.Now()

	for _, emoji := range emojis {
		go func(e Emoji) {
			defer wg.Done()
			err := downloadFile(e.Png, "emojis/"+e.Name+".png")
			if err != nil {
				fmt.Printf("Failed to download %s (%s): %v\n", e.Png, e.Name, err)
			}
		}(emoji)
	}

	wg.Wait()

	fmt.Println("Downloaded", len(emojis), "emojis in", time.Since(start))
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
