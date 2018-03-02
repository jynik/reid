/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * Search functionality
 */
package reid

import (
	"errors"
	"io/ioutil"
)

type searchFilter struct {
	config *procSearchConfig

	anyPublication bool
	publications   map[string]bool

	anyAuthor bool
	authors   map[string]bool
}

func (c *procSearchConfig) searchFilter() searchFilter {
	var f searchFilter

	f.anyPublication = (len(c.publications) == 0)
	if !f.anyPublication {
		f.publications = make(map[string]bool, len(c.publications))
		for _, p := range c.publications {
			f.publications[p] = true
		}
	}

	f.anyAuthor = (len(c.authors) == 0)
	if !f.anyAuthor {
		f.authors = make(map[string]bool, len(c.authors))
		for _, a := range c.authors {
			f.authors[a] = true
		}
	}

	return f
}

func (f *searchFilter) matches(e *ProjectEntry) bool {
	if !f.anyPublication {
		if !f.publications[Reduce(e.Record.Publication)] {
			return false
		}
	}

	if !f.anyAuthor {
		for _, author := range e.Record.Authors {
			author := Reduce(author)
			if f.authors[author] {
				return true
			}
		}

		return false
	}

	return true
}

func doSearch(c procSearchConfig, e *ProjectEntry) ([]SearchResult, error) {
	var results []SearchResult

	if len(e.MiniFiles) == 0 {
		Errorf("No converted text availble for: %s\n", e.Record.String())
		return []SearchResult{}, nil
	}

	Debugf("Searching: %s\n", e.Record.String())

	for _, filename := range e.MiniFiles {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return []SearchResult{}, err
		}
		Debugf("     Loaded: %s\n", filename)

		for i, query := range c.queries {
			Debugf("       Executing query %d of %d: \"%s\"\n", i+1, len(c.queries), query.orig)

			count := len(query.regexp.FindAllIndex(data, -1))
			Debugf("         Found %d occurrences\n", count)
			if count == 0 {
				continue
			}
			result := SearchResult{Query: query.orig, Occurrences: count, Record: e.Record}
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *Project) Search(s SearchConfig) ([]SearchResult, error) {
	var results []SearchResult

	config, err := s.process()
	if err != nil {
		return []SearchResult{}, err
	}

	filter := config.searchFilter()

	if s.Start > s.End {
		return []SearchResult{}, errors.New("Start year must be >= End year")
	}

	for year := s.Start; year <= s.End; year++ {
		entries, haveRecords := p.yearMap[year]
		if !haveRecords {
			continue
		}

		for _, entry := range entries {
			if filter.matches(entry) {
				Verbosef("Filter matched record: %s\n", entry.Record.String())
				r, err := doSearch(config, entry)
				if err != nil {
					return []SearchResult{}, err
				}

				results = append(results, r...)
			}
		}
	}

	return results, nil
}
