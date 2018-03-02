/*
 * Copyright (c) 2017-2018 Jon Szymaniak <jon.szymaniak@gmail.com>
 * SPDX License Identifier: GPL-3.0
 *
 */
package reid

import "fmt"

type VersionInfo struct {
	Major  uint
	Minor  uint
	Patch  uint
	Suffix string
}

func (v VersionInfo) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Suffix) != 0 {
		version += "-" + v.Suffix
	}

	return version
}
