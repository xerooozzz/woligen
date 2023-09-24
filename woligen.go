package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

func main() {
	var concurrency int
	var urlsFile string
	var parametersFile string
	flag.IntVar(&concurrency, "c", 30, "The concurrency for speed")
	flag.StringVar(&urlsFile, "u", "", "The file containing URLs")
	flag.StringVar(&parametersFile, "p", "", "The file containing parameters")
	flag.Parse()

	if urlsFile != "" && parametersFile != "" {
		urls, err := readLinesFromFile(urlsFile)
		if err != nil {
			fmt.Println("Error reading URLs file:", err)
			return
		}

		parameters, err := readLinesFromFile(parametersFile)
		if err != nil {
			fmt.Println("Error reading parameters file:", err)
			return
		}

		var wg sync.WaitGroup
		urlCh := make(chan string)

		// Start worker goroutines
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for urlStr := range urlCh {
					gen(urlStr, parameters)
				}
			}()
		}

		// Distribute URLs to worker goroutines
		for _, urlStr := range urls {
			urlCh <- urlStr
		}

		// Close the channel after all URLs are processed
		close(urlCh)

		// Wait for all worker goroutines to finish
		wg.Wait()
	}
}

func gen(urlStr string, parameters []string) {
	// Regex to find URLs
	URL_REGEX := `(?i)(?:(?:https?|ftp|smtp|unknown|sftp|file|data|telnet|ssh|ws|wss|git|svn|gopher):\/\/)(?:(?:[^\s:@'"]+(?::[^\s:@'"]*)?@)?(?:[_A-Z0-9.-]+|\[[_A-F0-9]*:[_A-F0-9:]+\])(?::\d{1,5})?)(?:\/[^\s'"]*)?(?:\?[^\s'"]*)?(?:#[^\s'"]*)?`

	time.Sleep(time.Millisecond * 10)

	// Find URLs in the text
	re := regexp.MustCompile(URL_REGEX)
	matches := re.FindAllString(urlStr, -1)
	if matches == nil {
		return
	}

	for _, match := range matches {
		parsedURL, err := url.Parse(match)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			continue
		}

		// Check if the URL already has parameters
		if parsedURL.RawQuery != "" {
			// Append parameters using "&" separator
			for _, parameter := range parameters {
				if parameter != "" {
					fmt.Printf("%s&%s=FUZZ\n", match, parameter)
				}
			}
		} else {
			// Append parameters using "?" separator
			for _, parameter := range parameters {
				if parameter != "" {
					fmt.Printf("%s?%s=FUZZ\n", match, parameter)
				}
			}
		}
	}
}

func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
