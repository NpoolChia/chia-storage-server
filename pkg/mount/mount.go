package mount

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/NpoolChia/chia-storage-server/util"
)

const (
	mountRoot = "/mnt"
)

type (
	mountInfo struct {
		// 挂载点
		path string
		// 大小
		size int64
		// .tmp count
		tmpFileCount int8
	}

	// all mount point info
	mountInfos []mountInfo
)

var (
	_mountInfos mountInfos
)

func (a mountInfos) Len() int      { return len(a) }
func (a mountInfos) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a mountInfos) Less(i, j int) bool {
	// first sort by size, then iorate
	return a[i].size < a[j].size
}

// Choose the right moint point
func (a mountInfos) mount() mountInfo {
	if len(a) > 0 {
		// lazy sort
		sort.Sort(mountInfos(a))
		return a[0]
	}
	return mountInfo{}
}

// Mount 寻找合适的目录
func Mount() string {
	return _mountInfos.mount().path
}

// InitMount find all mount info
func InitMount() error {
	filepath.WalkDir(mountRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			ok, err := util.IsMountPoint(d.Name())
			if err != nil {
				return err
			}
			if ok {
				finfo, err := d.Info()
				if err != nil {
					return err
				}
				_mountInfos = append(_mountInfos, mountInfo{
					size: finfo.Size(),
					path: path,
				})
			}
		}
		return nil
	})
	return nil
}
