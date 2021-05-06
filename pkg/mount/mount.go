package mount

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/NpoolChia/chia-storage-server/util"
)

const (
	mountRoot = "/mnt"
	// TmpFileExt temporary file Extension
	TmpFileExt = ".tmp"
)

type (
	mountInfo struct {
		// 挂载点
		path string
		// 大小
		size int64
		// .tmp count
		tmpFileCount int
	}

	// all mount point info
	mountInfos []mountInfo
)

var (
	_mountInfos mountInfos
	lock        sync.Mutex
)

func (a mountInfos) Len() int      { return len(a) }
func (a mountInfos) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a mountInfos) Less(i, j int) bool {
	// first sort by size, then tmp file count
	if a[i].size == a[j].size {
		return a[i].tmpFileCount < a[j].tmpFileCount
	}
	return a[i].size < a[j].size
}

// Choose the right moint point
func (a mountInfos) mount() mountInfo {
	// lazy check
	initMount()
	if len(a) > 0 {
		if !sort.IsSorted(mountInfos(a)) {
			// lazy sort
			sort.Sort(mountInfos(a))
		}
		return a[0]
	}
	return mountInfo{}
}

// Mount 寻找合适的目录
func Mount() string {
	lock.Lock()
	defer lock.Unlock()
	return _mountInfos.mount().path
}

// InitMount find all mount info
func InitMount() error {
	return initMount()
}

func initMount() error {
	// read all mount dir
	mountEntry, err := os.ReadDir(mountRoot)
	if err != nil {
		return err
	}

	mountPoints := make(map[string]mountInfo)
	for _, mountPoint := range mountEntry {
		isDir := mountPoint.IsDir()
		if isDir {
			absMountPath := filepath.Join(mountRoot, mountPoint.Name())
			ok, err := util.IsMountPoint(absMountPath)
			if err != nil {
				return err
			}
			if ok {
				// find all sub file, then statistics all file size
				filepath.Walk(absMountPath, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						tmpFile := 0
						if filepath.Ext(info.Name()) == TmpFileExt {
							tmpFile = 1
						}
						if v, ok := mountPoints[absMountPath]; ok {
							mountPoints[absMountPath] = mountInfo{
								path:         absMountPath,
								size:         v.size + info.Size(),
								tmpFileCount: v.tmpFileCount + tmpFile,
							}
						} else {
							mountPoints[absMountPath] = mountInfo{
								path:         absMountPath,
								size:         info.Size(),
								tmpFileCount: tmpFile,
							}
						}
					}
					return nil
				})
			}
		}
	}

	lock.Lock()
	_mountInfos = []mountInfo{}
	for _, v := range mountPoints {
		_mountInfos = append(_mountInfos, v)
	}
	lock.Unlock()

	return nil
}
