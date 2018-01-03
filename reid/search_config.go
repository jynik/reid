/*
* Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
* SPDX License Identifier: GPL-3.0
*
* Search configuration
*/
package reid

import (
	"fmt"
	"regexp"
	"strings"
)

// Caller-provided search configuration
type SearchConfig struct {
	Terms	[]string
	Regexps []string
	Authors	[]string
	Publications []string
	Start	int
	End		int
}

type query struct {
	orig	string			// Regexp pattern or search term
	regexp	*regexp.Regexp	// Compiled regular expression
}

// Processed search configuration
type procSearchConfig struct {
	queries	[]query
	authors []string
	publications []string
}

/* Process search configuration up front to avoid repeated
 * text minification or regexp compilation.
 */
func (s *SearchConfig) process() (procSearchConfig, error) {
	var proc procSearchConfig
	var err error

	proc.queries = make([]query, len(s.Regexps) + len(s.Terms))
	for i, pattern := range(s.Regexps) {
		Verbosef("Compiling regexp pattern \"%s\"\n", pattern)
		proc.queries[i].regexp, err = regexp.Compile(pattern)
		if err != nil {
			return procSearchConfig{}, err
		}
		proc.queries[i].orig = "regexp{" + pattern + "}"
	}

	// Build a simple regular expressions from search terms
	r := len(s.Regexps)
	for i, term := range(s.Terms) {
		// Strip text that won't be present in our minified text files
		t := reNonAlnumSpace.ReplaceAllString(strings.ToLower(term), "")

		// Merge excess whitespace into a single space
		t = reExtraSpace.ReplaceAllString(t, " ")

		// Match only complete words, delimited by whitespace or line endings
		pattern := "(^| )"+t+"( |$)"

		Verbosef("Converting search term \"%s\" -> regexp{%s}\n", term, pattern)
		proc.queries[r+i].regexp, err = regexp.Compile(pattern)
		if err != nil {
			return procSearchConfig{}, err
		}
		proc.queries[r+i].orig = t
	}

	proc.authors = make([]string, len(s.Authors))
	for i, author := range(s.Authors) {
		a := Reduce(author)
		if len(a) == 0 {
			return procSearchConfig{}, fmt.Errorf("Invalid author name: %s\n", author)
		}
		proc.authors[i] = a
	}

	proc.publications = make([]string, len(s.Publications))
	for i, pub := range(s.Publications) {
		p := Reduce(pub)
		if len(p) == 0 {
			return procSearchConfig{}, fmt.Errorf("Invalid publication name: %s\n", pub)
		}
		proc.publications[i] = p
	}

	return proc, nil
}
