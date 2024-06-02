package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

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
