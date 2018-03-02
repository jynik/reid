/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * Some simple sets for insertion and readback of unique values
 */
package reid

import "strings"

type IntSet struct {
	Values []int
	have   map[int]bool
}

func NewIntSet(capacity int) IntSet {
	var ret IntSet
	ret.Values = make([]int, 0, capacity)
	ret.have = make(map[int]bool, capacity)
	return ret
}

func (s *IntSet) Insert(val int) {
	if !s.have[val] {
		s.have[val] = true
		s.Values = append(s.Values, val)
	}
}

type StringSet struct {
	Values []string
	have   map[string]bool
}

func NewStringSet(capacity int) StringSet {
	var ret StringSet
	ret.Values = make([]string, 0, capacity)
	ret.have = make(map[string]bool, capacity)
	return ret
}

func (s *StringSet) Insert(val string) {
	if !s.have[val] {
		s.have[val] = true
		s.Values = append(s.Values, val)
	}
}

func (s *StringSet) CaseInsensitveInsert(val string) {
	lowerVal := strings.ToLower(val)
	if !s.have[lowerVal] {
		s.have[lowerVal] = true
		s.Values = append(s.Values, val)
	}
}

type RecordSet struct {
	Values []Record
	has    map[RecordHash]bool
}

func NewRecordSet(capacity int) RecordSet {
	var ret RecordSet
	ret.Values = make([]Record, 0, capacity)
	ret.has = make(map[RecordHash]bool, capacity)
	return ret
}

// Returns true if inserted, false if already in set
func (s *RecordSet) Insert(r *Record) bool {
	hash := r.Hash()
	if !s.has[hash] {
		s.has[hash] = true
		s.Values = append(s.Values, *r)
		return true
	} else {
		return false
	}
}
