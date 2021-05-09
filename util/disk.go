package util

import (
	syscall "golang.org/x/sys/unix"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type diskStatus struct {
	path  string
	all   uint64
	used  uint64
	free  uint64
	avail uint64
}

// DFer df -h
type DFer interface {
	GetAvail() (uint64, error)
	diskUsage() error
}

func New(path string) DFer {
	return &diskStatus{
		path: path,
	}
}

func (x *diskStatus) GetAvail() (uint64, error) {
	err := x.diskUsage()
	if err != nil {
		return 0, err
	}
	return x.avail, nil
}

// disk usage of path/disk
func (x *diskStatus) diskUsage() error {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(x.path, &fs)
	if err != nil {
		return err
	}
	// kb
	x.all = fs.Blocks * uint64(fs.Bsize)
	x.avail = fs.Bavail * uint64(fs.Bsize)
	x.free = fs.Bfree * uint64(fs.Bsize)
	x.used = x.all - x.free
	return nil
}
