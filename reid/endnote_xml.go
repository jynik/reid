/*
 * Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * EndNote XML "parsing". Developed against EndNode X7, and only scrapes
 * fields we're particularly interested in. This is not intended to be a
 * generic replacement for a proper XML parser.
 */
package reid

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type xmlLoader struct {
	file io.ReadCloser
	dec  *xml.Decoder
}

func newXmlLoader(filename string) (*xmlLoader, error) {
	var l *xmlLoader = new(xmlLoader)
	var err error

	l.file, err = os.Open(filename)
	if err != nil {
		return nil, err
	}

	l.dec = xml.NewDecoder(l.file)
	return l, nil
}

// Read and discard data until the specified start or end tag is found
// `name` is assumed to be lowercase
func (l *xmlLoader) seekToElt(name string, start, end bool) error {
	for {
		tok, err := l.dec.Token()
		if err != nil {
			return err
		}

		switch elt := tok.(type) {
		case xml.StartElement:
			if start && name == strings.ToLower(elt.Name.Local) {
				return nil
			}

		case xml.EndElement:
			if end && name == strings.ToLower(elt.Name.Local) {
				return nil
			}
		}
	}
}

// Read and discard data until the specified start tag is found
// `name` is assumed to be lowercase
func (l *xmlLoader) seekToStartElt(name string) error {
	return l.seekToElt(name, true, false)
}

// Read and discard data until the specified start tag is found
// `name` is assumed to be lowercase
func (l *xmlLoader) seekToEndElt(name string) error {
	return l.seekToElt(name, false, true)
}

// Skip everything until the end of the current record
func (l *xmlLoader) skipRecord() error {
	if err := l.seekToEndElt("record"); err != nil {
		return err
	} else {
		return errors.New("Skipped record due to error")
	}
}

func (l *xmlLoader) loadRecordDatabase(elt xml.StartElement) (string, error) {
	var path, name string

	Verbose("Processing database")
	for _, attr := range elt.Attr {
		if strings.ToLower(attr.Name.Local) == "path" {
			path = filepath.Dir(attr.Value)
		} else if strings.ToLower(attr.Name.Local) == "name" {
			name = attr.Value
		}
	}

	if len(path) == 0 {
		Debug("No database path in the current record. Skipping the rest of this record.")
		return "", l.skipRecord()
	} else if len(name) == 0 {
		Debug("No database name in the current record. Skipping the rest of this record.")
		return "", l.skipRecord()
	}

	return filepath.Join(path, strings.Replace(name, ".enl", "", -1)), nil
}

func (l *xmlLoader) readDataString(currElt string) (string, error) {
	var s string

	for {
		tok, err := l.dec.Token()
		if err != nil {
			return "", nil
		}

		switch elt := tok.(type) {
		case xml.CharData:
			s += string(elt)
		case xml.EndElement:
			if strings.ToLower(elt.Name.Local) == currElt {
				return s, nil
			}
		}
	}
}

func (l *xmlLoader) readDataInt(currElt string, min, max int) (int, error) {
	if s, err := l.readDataString(currElt); err != nil {
		return -1, err
	} else if val, err := strconv.Atoi(s); err != nil {
		return -1, err
	} else if val < min || val > max {
		return -1, errors.New("Integer outside of valid range")
	} else {
		return val, nil
	}
}

func (l *xmlLoader) readDataStrings(parent, target string) ([]string, error) {
	var strs []string
	var name string

	for {
		tok, err := l.dec.Token()
		if err != nil {
			return []string{}, err
		}

		switch elt := tok.(type) {
		case xml.StartElement:
			name = strings.ToLower(elt.Name.Local)
		case xml.EndElement:
			if strings.ToLower(elt.Name.Local) == parent {
				return strs, nil
			}
		default:
			continue
		}

		if name == target {
			Verbosef("Processing element: %s\n", target)
			s, err := l.readDataString(target)
			if err != nil {
				return []string{}, err
			}
			strs = append(strs, s)
			name = ""
		} else {
			Verbosef("Ignoring start element while processing %s: %s\n", target, name)
		}
	}
}

// TODO Only care about "title" and "publication" right now
func (l *xmlLoader) loadRecordTitles(r *Record) error {
	var eltName string

	Verbose("Processing titles")
	for {
		tok, err := l.dec.Token()
		if err != nil {
			return err
		}

		switch elt := tok.(type) {
		case xml.StartElement:
			eltName = strings.ToLower(elt.Name.Local)
		case xml.EndElement:
			if strings.ToLower(elt.Name.Local) == "titles" {
				return nil
			} else {
				continue
			}
		default:
			continue
		}

		switch eltName {
		case "title":
			r.Title, err = l.readDataString(eltName)
			if err != nil {
				return err
			}

		// I've been seeing these all inconsistently containing Journal titles or abbreviations.
		// As a hack, assume the longest one is actually the publication title.
		case "publication", "secondary-title", "alt-title", "full-title":
			pub, err := l.readDataString(eltName)
			if err != nil {
				return err
			}

			if len(pub) > len(r.Publication) {
				r.Publication = pub
			}

		default:
			continue
		}
	}
}

func (l *xmlLoader) loadRecord() (*Record, error) {
	var rec Record

	var dbPath string
	var pdfs []string

	var eltName string
	var startElt xml.StartElement

	Verbose("Processing record")

	for {
		tok, err := l.dec.Token()
		if err != nil {
			return nil, err
		}

		switch elt := tok.(type) {
		case xml.StartElement:
			eltName = strings.ToLower(elt.Name.Local)
			startElt = elt
		case xml.EndElement:
			if strings.ToLower(elt.Name.Local) == "record" {

				for _, pdf := range pdfs {
					// We're only concerned with local PDFS for now
					// TODO Track other URLS?
					internal := "internal-pdf://"
					if strings.HasPrefix(pdf, internal) {
						pdf = strings.Replace(pdf, internal, "", 1)

						// QueryUnescape will turn '+' in filenames into spaces,
						// which is not desirable. As hack, we'll replace it with a marker
						// and then swap it back out...
						pdf = strings.Replace(pdf, "+", "__REIDPLUS__", -1)

						// These are escaped in the XMLs I've seen so far
						pdf, err = url.QueryUnescape(pdf)
						if err != nil {
							return nil, err
						}

						// Reintroduce those crazy '+' characters...
						pdf = strings.Replace(pdf, "__REIDPLUS__", "+", -1)

						rec.PDFs = append(rec.PDFs, filepath.Join(dbPath + ".Data", "PDF", pdf))
					}
				}

				complete, missing := rec.isComplete()
				if complete && len(dbPath) == 0 {
					complete = false
					missing = "Database path"
				}

				// The <library>.DATA directory lives alongside the EndNote library
				dbPath = strings.Replace(dbPath, ".enl", ".DATA", 1)

				if !complete {
					if missing == "Title" {
						Debug("Not including incomplete record - missing Title\n")
					} else {
						Debugf("Not including incomplete record \"%s\" - missing %s\n",
							rec.Title, missing)
					}
					return nil, nil
				}

				return &rec, nil
			}
		default:
			continue // Data we're not interested in
		}

		// Need to consume corresponding xml.EndElement if the called fn does not do it for us
		needEndElt := true

		switch eltName {
		case "database":
			dbPath, err = l.loadRecordDatabase(startElt)

		case "contributors":
			/* TODO I've only seen authors listed in <contributors> so far.
			 *		What other types of contributors can be present here? */
			err = l.seekToStartElt("authors")
			if err != nil {
				return nil, err
			}

			Verbose("Processing authors")
			rec.Authors, err = l.readDataStrings("authors", "author")

		case "titles":
			needEndElt = false
			err = l.loadRecordTitles(&rec)

		case "language":
			needEndElt = false
			Verbosef("Processing %s\n", eltName)
			rec.Language, err = l.readDataString(eltName)

		case "dates":
			Verbosef("Processing %s\n", eltName)

			// We only care about the publication year
			err = l.seekToStartElt("year")
			if err == nil {
				rec.Year, err = l.readDataInt("year", 1, 3030)
				if err != nil {
					Debug("Invalid date encountered. Skipping the rest of the current record.")
					err = l.skipRecord()
				}
			}

		case "urls":
			Verbosef("Processing %s\n", eltName)
			needEndElt = false
			pdfs, err = l.readDataStrings("urls", "url")

		default:
			Verbosef("Ignoring element while processing record: %s\n", eltName)
		}

		if err != nil {
			return nil, err
		} else if needEndElt {
			err = l.seekToEndElt(eltName)
			if err != nil {
				return nil, err
			}
		}
	}
}

// Arbitrary "more than I ever expect to need" pre-allocation
const recSetCapacity = 10000

func LoadRecordsFromXML(filename string, filterLangs []string) ([]Record, error) {
	var records RecordSet = NewRecordSet(5000)
	var err error

	l, err := newXmlLoader(filename)
	if err != nil {
		return []Record{}, err
	}

	err = l.seekToStartElt("records")
	if err != nil {
		return []Record{}, err
	}

	// Perform case-insensitive language comparissons
	for i := 0; i < len(filterLangs); i++ {
		filterLangs[i] = strings.ToLower(filterLangs[i])
	}

	for {
		err = l.seekToStartElt("record")
		if err == io.EOF {
			return records.Values, nil
		} else if err != nil {
			return []Record{}, err
		}

		rec, err := l.loadRecord()
		if err != nil {
			return []Record{}, err
		} else if rec != nil {
			if rec.IsWrittenIn(filterLangs) {
				newRecord := records.Insert(rec)
				if !newRecord {
					Debug("Not including potential duplicate:", rec)
				} else {
					Debug("Loaded record: ", rec)
				}
			} else {
				Debugf("Not including due to Language=%s: %s\n", rec.Language, rec)
			}
		}
	}
}
