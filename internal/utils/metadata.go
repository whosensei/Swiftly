package utils

import (
    "fmt"
    "log"
    "net/http"
    "strings"
    "github.com/PuerkitoBio/goquery"
)

type Metadata struct {
    Title       string
    Description string
    ImageURL    string
    SiteName    string
}

func extractMetadata(url string) (*Metadata, error) {

	res, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to make HTTP request: %w", err)
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("bad status code: %d %s", res.StatusCode, res.Status)
    }

    // Parse the HTML document with goquery
    doc, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to parse HTML: %w", err)
    }

    metadata := &Metadata{}

    // Extract Open Graph (og:) properties
    doc.Find("meta").Each(func(i int, s *goquery.Selection) {
        property, exists := s.Attr("property")
        if !exists {
            return
        }

        content, contentExists := s.Attr("content")
        if !contentExists {
            return
        }

        switch property {
        case "og:title":
            if metadata.Title == "" {
                metadata.Title = content
            }
        case "og:description":
            if metadata.Description == "" {
                metadata.Description = content
            }
        case "og:image":
            if metadata.ImageURL == "" {
                metadata.ImageURL = content
            }
        case "og:site_name":
            if metadata.SiteName == "" {
                metadata.SiteName = content
            }
        }
    })

    // Fallback to standard meta tags if Open Graph tags are missing
    if metadata.Title == "" {
        metadata.Title = strings.TrimSpace(doc.Find("title").Text())
    }

    if metadata.Description == "" {
        doc.Find("meta[name='description']").Each(func(i int, s *goquery.Selection) {
            if content, exists := s.Attr("content"); exists {
                metadata.Description = content
            }
        })
    }
    return metadata, nil
}

func GetMetadata(url string) {

	favicon := fmt.Sprintf("https://www.google.com/s2/favicons?sz=64&domain_url=%s",url)
    metadata, err := extractMetadata(url)
    if err != nil {
        log.Fatalf("Error extracting metadata: %v", err)
    }

    fmt.Printf("Title: %s\n", metadata.Title)
    fmt.Printf("Description: %s\n", metadata.Description)
    fmt.Printf("Image URL (og:image): %s\n", metadata.ImageURL)
    fmt.Printf("Site Name: %s\n", metadata.SiteName)
	fmt.Printf("The favicon : %s\n",favicon)
}
