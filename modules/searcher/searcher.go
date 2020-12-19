package searcher

import (
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"strings"
)

// File holds the file contents in memory and a suffix array on which the search needs to be performed
type File struct {
	Content     string
	SuffixArray *suffixarray.Index
}

// Load loads the text file contents in memory
func (f *File) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}

	// Converting all the file contents to lowercase for handling case insensitive searches
	fileWithContentInLowerCase := strings.ToLower(string(dat))
	f.Content = fileWithContentInLowerCase

	f.SuffixArray = suffixarray.New([]byte(fileWithContentInLowerCase))

	return nil
}

// SearchString searches and returns substrings that have the search term in them
func (f *File) SearchString(query string) []string {
	searchStringLength := len(query)

	// Builds word occurence index
	indexPositions := f.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)

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
			if f.Content[i] == '.' || f.Content[i] == '\n' || f.Content[i] == '\r' {
				break
			}

			stepsBackward++
		}

		// Checks for the end of the sentence
		stepsForward := searchStringLength
		forwardStartPosition := indexPosition + searchStringLength

		for j := forwardStartPosition; ; j++ {
			if f.Content[j] == '.' || f.Content[j] == '\n' || f.Content[j] == '\r' {
				break
			}

			stepsForward++
		}

		// Adds sentence containing the searched word
		if stepsForward == searchStringLength {
			results = append(results, f.Content[indexPosition-stepsBackward:indexPosition+searchStringLength])
		} else {
			results = append(results, f.Content[indexPosition-stepsBackward:indexPosition+stepsForward])
		}
	}

	return results
}
