/*
* Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
* SPDX License Identifier: GPL-3.0
*
* Regular expressions used through the package
*/
package reid

import("regexp")

var reNonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var reNonAlnumSpace = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

/* The following regular expressions are used to strip text from the PDF
 * that will interfere with our searches */
var reUrl = regexp.MustCompile(`(https?://|www.)[a-zA-Z0-9./]+`) // TODO DOI
var reRefs = regexp.MustCompile(`\[[0-9]+\]`)
var rePunc = regexp.MustCompile(`[.:;,()]`)
var reQuotes = regexp.MustCompile(`["'‘’]`)
var reHyphenation = regexp.MustCompile(`-\s*\r?\n\s`)
var reNewlines = regexp.MustCompile(`\r?\n`)
var reExtraSpace = regexp.MustCompile(` +`)

