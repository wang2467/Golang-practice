package main

import (
	//"encoding/csv"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	//"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	var (
		urlTemplate = flag.String("url", "http://example.com/v/%d", "provide url")
		idLow       = flag.Int("from", 0, "the first id to scrape from")
		idHigh      = flag.Int("to", 0, "the last id to scrape")
		concurrency = flag.Int("concurrency", 0, "number of scrapers to work in parallel")
		name        = flag.String("nameQuery", "", "name")
		address     = flag.String("addressQuery", "", "address")
		phone       = flag.String("phoneQuery", "", "phone number")
		email       = flag.String("emailQuery", "", "email address")
	)
	flag.Parse()

	columns := []string{*name, *address, *phone, *email}
	headers := []string{"url", "id", "name", "address", "phone", "email"}

	type task struct {
		url string
		id  int
	}

	tasks := make(chan task)

	go func() {
		for i := *idLow; i < *idHigh; i++ {
			tasks <- task{url: fmt.Sprintf(*urlTemplate, i), id: i}
		}
		close(tasks)
	}()

	results := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(*concurrency)
	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()
			for n := range tasks {
				res, err := fetch(n.url, n.id, columns)
				if err != nil {
					log.Printf("could not fetch %s, %v", n.url, err)
					continue
				}
				results <- res
			}
		}()
	}

	fmt.Println(headers)
}

func fetch(url string, id int, queries []string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("Access limited")
		}
		return nil, fmt.Errorf("Bad server request %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse page: %v", err)
	}

	r := []string{url, strconv.Itoa(id)}
	for _, q := range queries {
		r = append(r, strings.TrimSpace(doc.Find(q).Text()))
	}
	return r, nil

}
