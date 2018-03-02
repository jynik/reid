/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * Search result object
 */

package reid

import (
	"fmt"
	"strings"
)

type SearchResult struct {
	Query       string
	Occurrences int
	Record      Record
}

func (r SearchResult) Pretty(eol string) string {
	return fmt.Sprintf(
		"Query: %s\n"+
			"   Occurrences: %d\n"+
			"   Year:        %d\n"+
			"   Publication: %s\n"+
			"   Author(s):   %s\n"+
			"   Title:       %s\n"+
			"%s%s",
		r.Query,
		r.Occurrences,
		r.Record.Year,
		r.Record.Publication,
		strings.Join(r.Record.Authors, " / "),
		r.Record.Title,
		eol, eol)
}

func (r SearchResult) PrettyBytes(eol string) []byte {
	return []byte(r.Pretty(eol))
}

func SearchResultCSVHeader(sep, eol string) string {
	return fmt.Sprintf(
		"Query%s"+
			"Occurrences%s"+
			"Year%s"+
			"Publication%s"+
			"Author(s)%s"+
			"Title%s",
		sep, sep, sep, sep, sep, eol)
}

func SearchResultCSVHeaderBytes(sep, eol string) []byte {
	return []byte(SearchResultCSVHeader(sep, eol))
}

func (r SearchResult) CSV(sep, eol string) string {
	authors := strings.Join(r.Record.Authors, " / ")
	return fmt.Sprintf(
		`"%s"%s`+ // Query
			`"%d"%s`+ // Occurrences
			`"%d"%s`+ // Year
			`"%s"%s`+ // Publication
			`"%s"%s`+ // Author(s)
			`"%s"%s`, // Title
		r.Query, sep,
		r.Occurrences, sep,
		r.Record.Year, sep,
		r.Record.Publication, sep,
		authors, sep,
		r.Record.Title, eol)
}

func (r SearchResult) CSVBytes(sep, eol string) []byte {
	return []byte(r.CSV(sep, eol))
}
