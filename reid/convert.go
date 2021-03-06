/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * Implementation of PDF -> minified text conversion
 */

package reid

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/gosseract/v1/gosseract"
)

/*
 * Specifies a record to convert from PDF to text.
 * The first field, in the order they are defined below, that matches a record
 * will be used.
 */
type RecordToConvert struct {
	Title       string
	Author      string
	Publication string
	Year        int
	Hash        string
}

func (r *RecordToConvert) String() string {
	return fmt.Sprintf("Title:\"%s\", Author:\"%s\", Publication:\"%s\" Year:%d Hash:\"%s\"",
		r.Title, r.Author, r.Publication, r.Year, r.Hash)
}

/*
 * Convert one or more PDF tiles to minified text files. If `records` contains
 * zero entries, all unconverted PDF files will be processed.
 *
 * If the `forceOCR` flag is set, the conversion will be perfomed using OCR, instead
 * of first trying to extract searchable text.
 *
 * If the `forceConversion` flag is specified, this will force specified
 * (including when "all" is implicitly by a zero-length `records`) PDFs to be
 * converted, even if a minified text file is already present.
 */
func (p *Project) Convert(records []RecordToConvert, forceOCR, forceConversion bool) error {
	if len(records) != 0 {
		return p.convertSubset(records, forceOCR, forceConversion)
	} else {
		var firstError error
		for i, _ := range p.Entries {
			err := p.convert(&p.Entries[i], forceOCR, forceConversion)
			if err != nil && firstError == nil {
				firstError = err
			}
		}

		// Update project file
		Verbosef("Saving project file: %s\n", p.filename)
		err := p.Save(p.filename)
		if err != nil && firstError == nil {
			firstError = err
		}

		return firstError
	}
}

func (p *Project) convertSubset(records []RecordToConvert, forceOCR, forceConversion bool) error {
	var firstError error
	entries, err := p.aggregateConversionList(records)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err := p.convert(entry, forceOCR, forceConversion)
		if err != nil && firstError == nil {
			// Report the first error we encounter, but keep chugging along
			firstError = err
		}
	}

	return firstError
}

// Helper type for aggregating conversion set
type convSet struct {
	p       *Project
	entries pEntryList
	loaded  map[*ProjectEntry]bool
}

func newConvSet(p *Project, capacity int) convSet {
	var c convSet
	c.loaded = make(map[*ProjectEntry]bool)
	return c
}

func (c *convSet) insertEntry(entry *ProjectEntry) bool {
	if entry == nil {
		return false
	}

	if c.loaded[entry] {
		Verbosef("Previously aggregated %s\n", entry.Record.String())
	} else {
		c.entries = append(c.entries, entry)
		c.loaded[entry] = true
		Verbosef("Aggregated record for conversion: %s\n", entry.Record.String())
	}

	return true
}

func (c *convSet) insert(e pEntryList) bool {
	if e == nil {
		return false
	}

	for _, entry := range e {
		if !c.insertEntry(entry) {
			return false
		}
	}

	return true
}

func (p *Project) aggregateConversionList(records []RecordToConvert) (pEntryList, error) {
	convSet := newConvSet(p, len(records))

	for _, record := range records {
		if len(record.Title) != 0 {
			title, err := NewReducedStr(record.Title)
			if err != nil {
				return pEntryList{}, nil
			} else if convSet.insert(p.titleMap[title.Reduced]) {
				continue
			}
		}

		if len(record.Author) != 0 {
			author, err := NewReducedStr(record.Author)
			if err != nil {
				return pEntryList{}, nil
			} else if convSet.insert(p.authMap[author.Reduced]) {
				continue
			}
		}

		if len(record.Publication) != 0 {
			pub, err := NewReducedStr(record.Publication)
			if err != nil {
				return pEntryList{}, nil
			} else if convSet.insert(p.pubMap[pub.Reduced]) {
				continue
			}
		}

		if record.Year != 0 {
			if convSet.insert(p.yearMap[record.Year]) {
				continue
			}
		}

		if len(record.Hash) != 0 {
			hash, err := StringToRecordHash(record.Hash)
			if err != nil {
				return pEntryList{}, nil
			} else if convSet.insertEntry(p.hashMap[hash]) {
				continue
			}
		}

		return pEntryList{},
			fmt.Errorf("Could not locate a record matching: %s\n", record.String())
	}

	return convSet.entries, nil
}

func (p *Project) convert(e *ProjectEntry, forceOCR, overwrite bool) error {
	var firstError error
	var miniFiles []string = make([]string, 0, len(e.Record.PDFs))
	for _, pdf := range e.Record.PDFs {
		miniFile, err := p.convertPDF(pdf, e, forceOCR, overwrite)
		if err != nil {
			Errorf("Failed to convert '%s' - %s\n", pdf, err)
			if firstError != nil {
				firstError = err
			}
		} else {
			miniFiles = append(miniFiles, miniFile)
		}
	}

	// Only record MiniFiles if all conversions suceeded
	if firstError == nil {
		if len(miniFiles) != 0 {
			Debugf("Successfully converted: %s\n", miniFiles)
		}
		e.MiniFiles = miniFiles
		Verbosef("Updated entry's MiniFiles: %s\n", e.MiniFiles)
	}
	return firstError
}

func pdfToImages(filename string) (string, error) {
	tmpDir, err := ioutil.TempDir("/tmp", "reid-convert-")
	if err != nil {
		return "", err
	}

	pfx := tmpDir + "/img"
	err = exec.Command("pdfimages", filename, pfx).Run()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return tmpDir, nil
}

var supportedExts = []string{".jpg", ".tif", ".tiff", ".pbm", ".ppm", ".png"}

func isSupportedImage(filename string) bool {
	ext := filepath.Ext(filename)
	for _, e := range supportedExts {
		if e == ext {
			return true
		}
	}
	return false
}

// TODO Extract this from the record to convert
const supportedLangs = "eng"

func imagesToText(dir string) (string, error) {
	var text string
	var files []string

	walk := func(filename string, info os.FileInfo, err error) error {
		if err == nil && isSupportedImage(filename) {
			files = append(files, filename)
		}
		return nil
	}
	filepath.Walk(dir, walk)

	for _, filename := range files {
		Debugf("Extracted text from %s\n", filename)
		text += gosseract.Must(gosseract.Params{Src: filename, Languages: supportedLangs})
		text += " "
	}

	return text, nil
}

// First try converting via searchable text. If this yields an empty
// file, we probably have a PDF that's scanned images -- attempt to use OCR.
//
// Returns minified output and error status
func (p *Project) convertAndMinify(filename string, forceOCR bool) ([]byte, error) {
	if !forceOCR {
		output, err := exec.Command("pdftotext", "-q", "-nopgbrk", "-enc", "UTF-8", "-eol", "unix", filename, "-").Output()
		if err != nil {
			return []byte{}, err
		}

		minText := minify(string(output))
		textLen := len(minText)
		if textLen == 0 {
			Debug("PDF did not contain searchable text. Using OCR conversion.")
		} else if textLen <= shortMiniTextThreshold {
			Debugf("Conversion yielded suspiciously low character count (%d). Trying OCR instead...\n", textLen)
		} else {
			Verbosef("Collected %d characters of searchable text.\n", textLen)
			return minText, nil
		}

	}

	imgDir, err := pdfToImages(filename)
	if err != nil {
		return []byte{}, err
	}
	defer os.RemoveAll(imgDir)

	text, err := imagesToText(imgDir)
	if err != nil {
		return []byte{}, err
	}

	return minify(text), nil
}

// Returns MiniFiles entry path, error
func (p *Project) convertPDF(filename string, e *ProjectEntry, forceOCR, overwrite bool) (string, error) {
	Infof("Converting %s\n", filename)

	pdf := filepath.Base(filename)
	subdir := filepath.Base(filepath.Dir(filename))

	targetDir := filepath.Join(p.DataDir, subdir)
	miniFile := filepath.Join(targetDir, pdf+".txt")

	// Ensure the requisite directory exists
	if err := os.MkdirAll(targetDir, 0770); err != nil {
		return "", err
	}

	// Only overwrite the file if requested
	if _, err := os.Stat(miniFile); !os.IsNotExist(err) && !overwrite {
		Debugf("%s already exists and an overwrite wasn't requested.\n", miniFile)
		return miniFile, nil
	}

	// Convert PDF->txt and minify it
	text, err := p.convertAndMinify(filename, forceOCR)
	if err != nil {
		return "", err
	}

	return miniFile, ioutil.WriteFile(miniFile, text, 0640)
}

func minify(text string) []byte {
	// Join hyphenation across lines
	text = reHyphenation.ReplaceAllString(text, "")

	// Replace newlines with whitespace
	text = reNewlines.ReplaceAllString(text, " ")

	// Remove references
	text = reRefs.ReplaceAllString(text, "")

	// Remove URLs
	text = reUrl.ReplaceAllString(text, "")

	// Remove punctuation that might get in our way
	text = rePunc.ReplaceAllString(text, "")

	// Remove quotes that can interfere with queries
	text = reQuotes.ReplaceAllString(text, "")

	// Remove excessive whitespace
	text = reExtraSpace.ReplaceAllString(text, " ")

	// Convert to lowercase
	text = strings.ToLower(text)

	// Return []bytes to for file write
	return []byte(text)
}
