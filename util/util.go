package util

import "strings"

func IsMountPoint(point string) (bool, error) {
	r, err := runCmd("mountpoint", point)
	if err != nil {
		return false, err
	}

	// check point is mount point
	if strings.HasSuffix(r, "is a mountpoint") {
		return true, nil
	}

	return true, nil
}
