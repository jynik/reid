/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * reid-search: Search for terms within converted records
 *
 * Run with --help for usage information.
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"../reid"
	c "./common"
)

type ResultFormat int

const (
	FormatInvalid = -1
	FormatPretty  = iota
	FormatCSV
	FormatHeaderlessCSV
	FormatJSON
)

var (
	projectFile = kingpin.
			Flag(c.FLAG_PROJECT, c.FLAG_PROJECT_DESC).
			Short(c.FLAG_PROJECT_SHORT).
			Required().
			String()

	debug   = kingpin.Flag(c.FLAG_DEBUG, c.FLAG_DEBUG_DESC).Bool()
	verbose = kingpin.Flag(c.FLAG_VERBOSE, c.FLAG_VERBOSE_DESC).Bool()
	version = kingpin.Flag(c.FLAG_VERSION, c.FLAG_VERSION_DESC).Bool()
)

func validateFormat(s string) ResultFormat {
	switch strings.ToLower(s) {
	case "csv":
		return FormatCSV
	case "csv-no-hdr":
		return FormatHeaderlessCSV
	case "pretty":
		return FormatPretty
	case "json":
		return FormatJSON
	default:
		return FormatInvalid
	}
}

func writeJSONResults(results []reid.SearchResult, outfile *os.File) error {
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func main() {
	var searchConfig reid.SearchConfig
	var formatStr string
	var outfilename string
	var outfile *os.File
	var err error
	var csvSep = ","
	var eol = "\n"

	kingpin.
		Flag("term", "Search for the specified term or phrase. "+
			"May be specified multiple times.").
		Short('t').
		StringsVar(&searchConfig.Terms)

	kingpin.
		Flag("regexp", "Search for matches to the provided regular expression. "+
			"May be specified multiple times.").
		Short('r').
		StringsVar(&searchConfig.Regexps)

	kingpin.
		Flag("from", "Inclusive lower range on years to search").
		Short('F').
		Default("1").
		IntVar(&searchConfig.Start)

	kingpin.
		Flag("to", "Inclusive upper range on years to search").
		Short('T').
		Default("3030").
		IntVar(&searchConfig.End)

	kingpin.
		Flag("author", "Limit search to specified author. "+
			"May be specified multiple times to expand search to multiple authors.").
		Short('a').
		StringsVar(&searchConfig.Authors)

	kingpin.
		Flag("publication", "Limit search to specified publication. "+
			"May be specified multiple times to expand search to multiple publications.").
		Short('P').
		StringsVar(&searchConfig.Publications)

	kingpin.
		Flag("format", "Format of results. Options are: pretty, csv, csv-no-hdr").
		Short('f').
		Default("pretty").
		StringVar(&formatStr)

	kingpin.
		Flag("outfile", "File to write results to. "+
			"Standard out is used if this is not specified.").
		Short('o').
		StringVar(&outfilename)

	c.ParseCommandLine()

	format := validateFormat(formatStr)
	if format == FormatInvalid {
		fmt.Fprintf(os.Stderr, "Invalid result format: %s\n", formatStr)
		os.Exit(1)
	}

	if *verbose {
		reid.LogLevel = reid.LogLevelVerbose
	} else if *debug {
		reid.LogLevel = reid.LogLevelDebug
	}

	project, err := reid.LoadProject(*projectFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	results, err := project.Search(searchConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}

	if len(outfilename) != 0 {
		outfile, err = os.Create(outfilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
	} else {
		outfile = os.Stdout
	}
	defer outfile.Close()

	switch format {
	case FormatCSV:
		outfile.Write(reid.SearchResultCSVHeaderBytes(csvSep, eol))
	case FormatJSON:
		err := writeJSONResults(results, outfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(5)
		}
		os.Exit(0)
	}

	for _, result := range results {
		var data []byte

		switch format {
		case FormatPretty:
			data = result.PrettyBytes(eol)
		case FormatCSV:
			data = result.CSVBytes(csvSep, eol)
		}

		outfile.Write(data)
	}
}
