/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * reid-enxml: Load and extract data from EndNote XML files
 *
 * Run with --help for usage information.
 */
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"../reid"
	c "./common"
)

const (
	CMD_SHOW      = "show"
	CMD_SHOW_DESC = "Display records or attributes extracted from the " +
		"provided XML file."

	ARG_SHOW      = "attr"
	ARG_SHOW_DESC = "Specify \"all\" to show all records, or one of the " +
		"following to list only specific attributes: " +
		"Title, Publication, Year, Author, Language, PDF"

	CMD_CREATE      = "create"
	CMD_CREATE_DESC = "Create a project file for use with reid-convert and reid-search."

	ARG_CREATE_PROJ      = "project"
	ARG_CREATE_PROJ_DESC = "Project file to create."

	ARG_CREATE_DIR      = "dir"
	ARG_CREATE_DIR_DESC = "Directory to store project files in. This " +
		"will be created if it does not already exist."
)

// Command-line configuration items
var (
	force = kingpin.
		Flag("force", "Force file overwrite if project file specified "+
			"to \"create\" command already exists.").
		Short('f').
		Bool()

	xmlFile = kingpin.
		Flag("xml", "EndNote XML file to load").
		Short('x').
		Required().
		String()

	langs = kingpin.
		Flag("lang", "Filter records (inclusively) based upon language.").
		Default("eng").
		Strings()

	debug   = kingpin.Flag(c.FLAG_DEBUG, c.FLAG_DEBUG_DESC).Bool()
	verbose = kingpin.Flag(c.FLAG_VERBOSE, c.FLAG_VERBOSE_DESC).Bool()
	version = kingpin.Flag(c.FLAG_VERSION, c.FLAG_VERSION_DESC).Bool()

	// show <title|publication|year|author|language|pdf>
	cmdShow = kingpin.Command(CMD_SHOW, CMD_SHOW_DESC)
	argShow = cmdShow.Arg(ARG_SHOW, ARG_SHOW_DESC).Required().String()

	// create <project file> <directory>
	cmdCreate        = kingpin.Command(CMD_CREATE, CMD_CREATE_DESC)
	argCreateProject = cmdCreate.Arg(ARG_CREATE_PROJ, ARG_CREATE_PROJ_DESC).Required().String()
	argCreateDir     = cmdCreate.Arg(ARG_CREATE_DIR, ARG_CREATE_DIR_DESC).Required().String()
)

func showAll(records []reid.Record) {
	for _, record := range records {
		fmt.Printf("Title: %s\nAuthor(s): %s\nPublication: %s\n"+
			"Year: %d\nLanguage: %s\nMetadata Hash: %s\n\n",
			record.Title, strings.Join(record.Authors, " / "),
			record.Publication, record.Year, record.Language,
			record.HashString())
	}
}

func showYears(records []reid.Record) {
	var yearSet = reid.NewIntSet(100)
	for _, rec := range records {
		yearSet.Insert(rec.Year)
	}

	sort.Ints(yearSet.Values)
	for _, val := range yearSet.Values {
		fmt.Println(val)
	}
}

func showPublications(records []reid.Record) {
	var pubSet = reid.NewStringSet(500)
	for _, rec := range records {
		pubSet.CaseInsensitveInsert(rec.Publication)
	}

	sort.Strings(pubSet.Values)
	for _, val := range pubSet.Values {
		fmt.Println(val)
	}
}

func showTitles(records []reid.Record) {
	var titleSet = reid.NewStringSet(5000)
	for _, rec := range records {
		titleSet.CaseInsensitveInsert(rec.Title)
	}

	sort.Strings(titleSet.Values)
	for _, val := range titleSet.Values {
		fmt.Println(val)
	}
}

func showAuthors(records []reid.Record) {
	var authorSet = reid.NewStringSet(10000)
	for _, rec := range records {
		for _, author := range rec.Authors {
			authorSet.CaseInsensitveInsert(author)
		}
	}

	sort.Strings(authorSet.Values)
	for _, val := range authorSet.Values {
		fmt.Println(val)
	}
}

func showLanguages(records []reid.Record) {
	var langSet = reid.NewStringSet(10)
	for _, rec := range records {
		langSet.CaseInsensitveInsert(rec.Language)
	}

	sort.Strings(langSet.Values)
	for _, val := range langSet.Values {
		if len(val) == 0 {
			val = "<Not Specified>"
		}
		fmt.Println(val)
	}
}

func showPDFs(records []reid.Record) {
	var pdfSet = reid.NewStringSet(5000)
	for _, rec := range records {
		for _, pdf := range rec.PDFs {
			pdfSet.Insert(pdf)
		}
	}

	sort.Strings(pdfSet.Values)
	for _, val := range pdfSet.Values {
		fmt.Println(val)
	}
}

type showFunc func(records []reid.Record)

func main() {
	var err error
	var show showFunc
	var records []reid.Record

	cmd := c.ParseCommandLine()

	if *verbose {
		reid.LogLevel = reid.LogLevelVerbose
	} else if *debug {
		reid.LogLevel = reid.LogLevelDebug
	}

	switch cmd {
	case CMD_SHOW:
		switch strings.ToLower(*argShow) {
		case "", "all":
			show = showAll
		case "year", "years":
			show = showYears
		case "pub", "publication", "publications":
			show = showPublications
		case "title", "titles":
			show = showTitles
		case "author", "authors":
			show = showAuthors
		case "lang", "lanuage", "languages":
			show = showLanguages
		case "pdf", "pdfs":
			show = showPDFs
		default:
			fmt.Fprintf(os.Stderr,
				"Invalid attribute (%s). Run \"help show\" for "+
					"more information.\n", *argShow)
			os.Exit(2)
		}

		if records, err = reid.LoadRecordsFromXML(*xmlFile, *langs); err == nil {
			show(records)
		}

	case CMD_CREATE:
		if _, err := os.Stat(*argCreateProject); !os.IsNotExist(err) && !*force {
			fmt.Fprintf(os.Stderr, "Error: %s already exists. Run with -f if "+
				"you want to overwrite it.\n", *argCreateProject)
			os.Exit(3)
		}

		if records, err = reid.LoadRecordsFromXML(*xmlFile, *langs); err == nil {
			if project, err := reid.NewProject(*argCreateDir, records); err == nil {
				err = project.Save(*argCreateProject)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Invalid command: %s\n", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
