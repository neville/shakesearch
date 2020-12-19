package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

// Searcher holds the file contents in memory on which the search needs to be performed
type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

// Load loads the text file contents in memory
func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}

	// Converting all the file contents to lowercase for handling case insensitive searches
	fileWithContentInLowerCase := strings.ToLower(string(dat))
	s.CompleteWorks = fileWithContentInLowerCase

	s.SuffixArray = suffixarray.New([]byte(fileWithContentInLowerCase))

	return nil
}

// Search searches and returns substrings that have the search term in them
func (s *Searcher) Search(query string) []string {
	searchStringLength := len(query)

	// Builds word occurence index
	indexPositions := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)

	results := []string{}
	for _, indexPosition := range indexPositions {
		// Checks for the beginning of the sentence
		stepsBackward := 0
		backwardsStartPosition := indexPosition - 1

		// Handle case where text starts with the searched word
		if backwardsStartPosition < 0 {
			backwardsStartPosition = 0
		}

		for i := backwardsStartPosition; ; i-- {
			if s.CompleteWorks[i] == '.' || s.CompleteWorks[i] == '\n' || s.CompleteWorks[i] == '\r' {
				break
			}

			stepsBackward++
		}

		// Checks for the end of the sentence
		stepsForward := searchStringLength
		forwardStartPosition := indexPosition + searchStringLength

		for j := forwardStartPosition; ; j++ {
			if s.CompleteWorks[j] == '.' || s.CompleteWorks[j] == '\n' || s.CompleteWorks[j] == '\r' {
				break
			}

			stepsForward++
		}

		// Adds sentence containing the searched word
		if stepsForward == searchStringLength {
			results = append(results, s.CompleteWorks[indexPosition-stepsBackward:indexPosition+searchStringLength])
		} else {
			results = append(results, s.CompleteWorks[indexPosition-stepsBackward:indexPosition+stepsForward])
		}
	}

	return results
}
