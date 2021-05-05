package util

func IsMountPoint(point string) (bool, error) {
	r, err := runCmd("mountpoint", point)
	if err != nil {
		return false, err
	}

	if r == 0 {
		return true, nil
	}

	return false, nil
}
