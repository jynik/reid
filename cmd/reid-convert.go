/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * reid-convert: Convert PDFs (specified in a reid project file) into
 * plaintext files that can be searched using reid-search.
 *
 * By default, this will convert all (currently unconverted) PDFs. Flags can be
 * used to convert a specific set of records, or to force records to be
 * reconverted.
 *
 * Run with --help for usage information.
 */
package main

import (
	"fmt"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"../reid"
	c "./common"
)

var records []reid.RecordToConvert

// Command-line configuration items
var (
	projectFile = kingpin.
			Flag(c.FLAG_PROJECT, c.FLAG_PROJECT_DESC).
			Short(c.FLAG_PROJECT_SHORT).
			Required().
			String()

	ocr = kingpin.
		Flag("ocr",
			"Force the use of OCR when converting a document. Use this if "+
				"you find that a document contains unsearchable text. This will "+
				"ensure reid-convert attempts to scrape all text, rather than "+
				"just a cover page or copyright statement prepended to the "+
				"actual material.").
		Short('o').
		Bool()

	force = kingpin.
		Flag("force",
			"Force conversion of a record, even if minified text already "+
				"exists for it.").
		Short('f').
		Bool()

	titles = kingpin.
		Flag("title",
			"Specify a project file entry to convert, by title. "+
				"Multiple titles may be specified. May be used in "+
				"conjunction with other entry specifier flags.").
		Short('t').
		Strings()

	authors = kingpin.
		Flag("author",
			"Specify a project file entry to convert, by author. "+
				"Multiple authors may be specified. May be used in "+
				"conjunction with other entry specifier flags.").
		Short('a').
		Strings()

	publications = kingpin.
			Flag("publication",
			"Specify a project file entry to convert, by publication. "+
				"Multiple publicatinos may be specified. May be used in "+
				"conjunction with other entry specifier flags.").
		Short('P').
		Strings()

	years = kingpin.
		Flag("year",
			"Specify a project file entry to convert, by year. "+
				"Multiple years may be specified. May be used in "+
				"conjunction with other entry specifier flags.").
		Short('y').
		Ints()

	hashes = kingpin.
		Flag("hash",
			"Specify a project file entry to convert, by metadata hash. "+
				"Multiple hashes may be specified. May be used in "+
				"conjunction with other entry specifier flags.").
		Short('H').
		Strings()

	debug   = kingpin.Flag(c.FLAG_DEBUG, c.FLAG_DEBUG_DESC).Bool()
	verbose = kingpin.Flag(c.FLAG_VERBOSE, c.FLAG_VERBOSE_DESC).Bool()
	version = kingpin.Flag(c.FLAG_VERSION, c.FLAG_VERSION_DESC).Bool()
)

func main() {
	c.ParseCommandLine()

	if *verbose {
		reid.LogLevel = reid.LogLevelVerbose
	} else if *debug {
		reid.LogLevel = reid.LogLevelDebug
	}

	for _, title := range *titles {
		records = append(records, reid.RecordToConvert{Title: title})
	}

	for _, author := range *authors {
		records = append(records, reid.RecordToConvert{Author: author})
	}

	for _, publication := range *publications {
		records = append(records, reid.RecordToConvert{Publication: publication})
	}

	for _, year := range *years {
		records = append(records, reid.RecordToConvert{Year: year})
	}

	for _, hash := range *hashes {
		records = append(records, reid.RecordToConvert{Hash: hash})
	}

	project, err := reid.LoadProject(*projectFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	err = project.Convert(records, *ocr, *force)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}
}
