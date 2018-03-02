/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 * In (quick and dirty) attempt to be resilient to associated (or duplicate)
 * EndNote entries differing in case, whitespace, punctuation, or other weird
 * stuff [like brackets] in fields we work with some "reduced" strings when
 * doing lookups.
 *
 * XXX This has some potential to be a footgun, however. Keep this in mind if
 *	   it seems there are some entry hash collisions.
 */

package reid

import (
	"fmt"
	"strings"
)

type ReducedStr struct {
	String  string // Original string
	Reduced string // Lowercase and alnum only
}

func NewReducedStr(s string) (ReducedStr, error) {
	var r ReducedStr
	var err error

	r.String = s
	r.Reduced = Reduce(s)

	if len(r.String) != 0 && len(r.Reduced) == 0 {
		err = fmt.Errorf("Reducing \"%s\" results in an empty string\n", s)
	}

	return r, err
}

// Stip a string of all non-alnum characters and convert to lowercase
func Reduce(s string) string {
	return reNonAlnum.ReplaceAllString(strings.ToLower(s), "")
}
