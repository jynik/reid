/*
 * Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 */
package reid

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

/*
 * A "Record" is a collection of metadata associated with a publication, along
 * with one or more paths to its full text PDF(s).
 */
type Record struct {
	PDFs []string // Path to one or more PDFs

	Title       string // Record title (e.g., article name)
	Publication string // Publication title (e.g., journal name)
	Year        int    // Publication year

	Authors []string // List of authors

	Language string // Language the text is in
}

// If record is complete (at load-time), returns: true, ""
// Otherwise, returns: false, <Name of first incomplete field>
func (r *Record) isComplete() (bool, string) {

	if len(r.Title) == 0 {
		return false, "Title"
	}

	if len(r.Publication) == 0 {
		return false, "Publication"
	}

	// FIXME Won't some fields will use materials dated B.C. ?
	if r.Year <= 0 {
		return false, "Year"
	}

	if len(r.Authors) == 0 {
		return false, "Authors"
	}

	if len(r.PDFs) == 0 {
		return false, "PDFs"
	}

	// Language is permitted to be empty - treated as a catch-all
	if len(r.Language) == 0 {
		Debug("No language specified for:", r)
	}

	return true, ""
}

func (r *Record) String() string {
	return fmt.Sprintf("\"%s\" %s (%s %d)", r.Title, r.Authors, r.Publication, r.Year)
}

func reduceString(s string) string {
	return reNonAlnum.ReplaceAllString(strings.ToLower(s), "")
}

func (r *Record) computeHash() []byte {
	title := reduceString(r.Title)
	pub := reduceString(r.Publication)
	fieldSep := byte('|')

	year := make([]byte, 2)
	binary.LittleEndian.PutUint16(year, uint16(r.Year))

	h := md5.New()
	h.Write([]byte(title))
	h.Write([]byte{fieldSep})
	h.Write([]byte{pub[0]})
	h.Write([]byte{fieldSep})
	h.Write(year)
	h.Write([]byte{fieldSep})
	h.Write([]byte{uint8(len(r.Authors))})
	h.Write([]byte{fieldSep})

	return h.Sum(nil)
}

/* Compute a hash over the Record's Title (case and punctuation insensitive),
 * Year, # of Authors, and first letter of the associated publication name.
 * This is intended to help identify duplicates in a library in cases where the
 * capitalization of fields or the abbreviation of items varies.
 */
func (r *Record) Hash() RecordHash {
	var recordHash RecordHash
	copy(recordHash[:], r.computeHash())
	return recordHash
}

// Same as Hash() but return a string representation
func (r *Record) HashString() string {
	return hex.EncodeToString(r.computeHash())
}

// Returns true if record matches one of the provided languages
// Assumes entries in `langs` are lower-case
func (r *Record) IsWrittenIn(langs []string) bool {
	if len(r.Language) == 0 || len(langs) == 0 {
		// Err on the side of processing false positives if none specified
		return true
	}

	recLang := strings.ToLower(r.Language)

	for _, lang := range langs {
		if recLang == lang {
			return true
		}
	}

	return false
}

/*
 * A RecordHash is the MD5 hash of its ("reduced") metadata fields
 */
type RecordHash [16]byte

func (h RecordHash) String() string {
	return hex.EncodeToString([]byte(h[:]))
}

func StringToRecordHash(s string) (RecordHash, error) {
	var ret RecordHash
	if hash, err := hex.DecodeString(s); err != nil {
		return ret, err
	} else {
		copy(ret[:], hash)
		return ret, nil
	}
}
