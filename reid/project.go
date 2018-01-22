/*
* Copyright (c) 2017 Jon Szymaniak <jon.szymaniak@gmail.com>
* SPDX License Identifier: GPL-3.0
*
* A reid project file
*/

package reid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	filename	string

	CreatedAt   string
	ReidVersion string
	DataDir     string
	Entries     []ProjectEntry

	hashes  []RecordHash
	hashMap map[RecordHash]*ProjectEntry

	years   []int
	yearMap map[int]pEntryList

	titles   []ReducedStr
	titleMap map[string]pEntryList // map[reducedStr.reduced]pEntryList

	pubs   []ReducedStr
	pubMap map[string]pEntryList // map[reducedStr.reduced]pEntryList

	auths   []ReducedStr
	authMap map[string]pEntryList // map[reducedStr.reduced]pEntryList

}

type ProjectEntry struct {
	Record    Record   // Record extracted from EndNote
	Hash      string   // Record Hash, used to identify record
	MiniFiles []string // Minified text files used for searching
}

type pEntryList []*ProjectEntry

func NewProject(dataDir string, records []Record) (*Project, error) {
	project := new(Project)

	project.CreatedAt = time.Now().String()
	project.ReidVersion = Version.String()

	if !filepath.IsAbs(dataDir) {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dataDir = filepath.Join(pwd, dataDir)
	}

	project.DataDir = dataDir

	for _, record := range records {
		project.Entries = append(project.Entries, ProjectEntry{Record: record, Hash: record.HashString(), MiniFiles: []string{}})
	}

	return project, nil
}

func (p *Project) Save(filename string) error {
	var err error

	projectPath := filepath.Dir(filename)

	// Ensure project path and data directory exist
	if err = os.MkdirAll(projectPath, 0770); err != nil {
		Debugf("Failed to create project path: %s\n", err)
		return err
	}

	if err = os.MkdirAll(p.DataDir, 0770); err != nil {
		Debugf("Failed to create data directory: %s\n", err)
		return err
	}

	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()

	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")

	p.CreatedAt = time.Now().String()
	return enc.Encode(p)
}

func LoadProject(filename string, skipConverted bool) (*Project, error) {
	var project = new(Project)
	project.filename = filename

	infile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	dec := json.NewDecoder(infile)
	err = dec.Decode(project)
	if err != nil {
		return nil, err
	}

	return project.scan(skipConverted)
}

// Sanity check the state of the project, report and drop bad entries,
// and populate look-up tables
func (p *Project) scan(skipConverted bool) (*Project, error) {

	// Does our data dir exist?
	if _, err := os.Stat(p.DataDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Data directory does not exist: %s\n", p.DataDir)
		} else {
			return nil, err
		}
	}

	p.hashes = make([]RecordHash, 0, len(p.Entries))
	p.hashMap = make(map[RecordHash]*ProjectEntry, len(p.Entries))

	// TODO Clearly an excessive pre-allocation. Drop this a bit?
	p.yearMap = make(map[int]pEntryList, 100)
	p.years = make([]int, 0, 100)

	p.titleMap = make(map[string]pEntryList, len(p.Entries))
	p.titles = make([]ReducedStr, 0, len(p.Entries))

	/* Arbitrary guesstimate of 6 authors per paper. Lame attempt at
	 * overallocating up front and reducing allocations later. */
	p.authMap = make(map[string]pEntryList, 6*len(p.Entries))
	p.auths = make([]ReducedStr, 0, 6*len(p.Entries))

	// Again - over-allocating p.pubs here
	p.pubMap = make(map[string]pEntryList, len(p.Entries))
	p.pubs = make([]ReducedStr, 0, len(p.Entries))

	for i, entry := range p.Entries {
		// Do we already have minified text files for this record?
		if !skipConverted && len(entry.MiniFiles) != 0 {
			haveAll := true
			for _, f := range entry.MiniFiles {
				if _, err := os.Stat(f); err != nil {
					haveAll = false
					break
				}
				continue
			}

			if haveAll {
				Debugf("Skipping record because MiniFiles aleady exist for { %s }\n", entry.Record)
				continue
			}
		}

		// Do the PDFs exist?
		skip := false
		for _, pdf := range entry.Record.PDFs {
			if _, err := os.Stat(pdf); err != nil {
				if os.IsNotExist(err) {
					Errorf("PDF does not exist: %s\n", pdf)
					Debugf(" `- Skipping Record: %s\n", entry.Record.String())
					skip = true
					break
				}
			}
		}

		if skip {
			continue
		} else {
			Verbosef("Loading entry for %s\n", entry.Record.String())
		}

		hash := entry.Record.Hash()
		p.hashMap[hash] = &p.Entries[i]
		p.hashes = append(p.hashes, hash)
		Verbosef("Added entry to hashMap[%s]\n", hash.String())

		year := entry.Record.Year
		if recordsByYear, exists := p.yearMap[year]; exists {
			p.yearMap[year] = append(recordsByYear, &p.Entries[i])
			Verbosef("Appended entry to yearMap: %d\n", year)
		} else {
			p.yearMap[year] = pEntryList{&p.Entries[i]}
			p.years = append(p.years, year)
			Verbosef("Created entry in new yearMap: %d\n", year)
		}

		title, err := NewReducedStr(entry.Record.Title)
		if err != nil {
			return nil, err
		}
		if recordsByTitle, exists := p.titleMap[title.Reduced]; exists {
			p.titleMap[title.Reduced] = append(recordsByTitle, &p.Entries[i])
			Verbosef("Appended entry to titleMap: %s\n", title)
		} else {
			p.titleMap[title.Reduced] = pEntryList{&p.Entries[i]}
			p.titles = append(p.titles, title)
			Verbosef("Created entry in new titleMap: %s\n", title.Reduced)
		}

		for _, auth := range entry.Record.Authors {
			auth, err := NewReducedStr(auth)
			if err != nil {
				return nil, err
			}

			if recordsByAuth, exists := p.authMap[auth.Reduced]; exists {
				p.authMap[auth.Reduced] = append(recordsByAuth, &p.Entries[i])
				Verbosef("Appended entry to authMap: %s\n", auth.Reduced)
			} else {
				p.authMap[auth.Reduced] = pEntryList{&p.Entries[i]}
				p.auths = append(p.auths, auth)
				Verbosef("Created entry in new authMap: %s\n", auth.Reduced)
			}
		}

		pub, err := NewReducedStr(entry.Record.Publication)
		if err != nil {
			return nil, err
		}

		if recordsByPub, exists := p.pubMap[pub.Reduced]; exists {
			p.pubMap[pub.Reduced] = append(recordsByPub, &p.Entries[i])
			Verbosef("Appended entry to pubMap: %s\n", pub.Reduced)
		} else {
			p.pubMap[pub.Reduced] = pEntryList{&p.Entries[i]}
			p.pubs = append(p.pubs, pub)
			Verbosef("Created entry in new pubMap: %s\n", pub.Reduced)
		}
	}

	return p, nil
}

func (p *Project) Years() []int {
	var years []int = make([]int, len(p.years))
	for i, year := range p.years {
		years[i] = year
	}
	return years
}

func (p *Project) Authors() []string {
	var authors []string = make([]string, len(p.auths))
	for i, author := range p.auths {
		authors[i] = author.String
	}
	return authors
}

func (p *Project) Publications() []string {
	var publications []string = make([]string, len(p.pubs))
	for i, publication := range p.pubs {
		publications[i] = publication.String
	}
	return publications
}

func (p *Project) Hashes() []string {
	var hashes []string = make([]string, len(p.hashes))
	for i, hash := range p.hashes {
		hashes[i] = hash.String()
	}
	return hashes
}
