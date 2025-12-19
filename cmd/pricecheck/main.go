package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LehaAlexey/Parsing/internal/parser"
)

func main() {
	urls := []string{
		"https://www.pech.ru/catalog/elektroochagi/elektricheskiy-kamin-electrolux-sphere-plus-efp-p-2720rls/",
		"https://ru.aircraft24.com/singleprop/beechcraft/55-baron-project--xi142530.htm",
		"https://sunseeker-russia.com/yacht/sunseeker-manhattan-66-017/",
		"https://www.dns-shop.ru/product/b30662bca87cd21a/girlanda-govee-curtain-light/",
		"https://5ka.ru/product/nektar-global-village-ananasovyy-950ml--3634676/",
		"https://book24.ru/product/ohota-na-ohotnika-8751063/",
		"https://www.santehnica.ru/product/375887.html",
	}

	ctx := context.Background()
	fetcher := parser.NewFetcher(parser.FetcherConfig{
		UserAgent: "Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	})
	extractor := parser.NewExtractor()

	for _, url := range urls {
		fmt.Println("URL:", url)

		body, _, err := fetcher.Fetch(ctx, url)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}

		price, cur, ok := extractor.Extract(body)
		if !ok {
			fmt.Println("result: price not found")
			continue
		}

		fmt.Printf("result: price=%d currency=%q\n\n", price, cur)
	}
}
