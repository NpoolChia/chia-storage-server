package mount

import (
	"os"
	"os/exec"
	"path/filepath"
	_ "sort"
	"sync"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
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
	_mountInfos   mountInfos
	lock          sync.Mutex
	curMountIndex int
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
	/*
		if len(a) > 0 {
			if !sort.IsSorted(mountInfos(a)) {
				// lazy sort
				sort.Sort(mountInfos(a))
			}
			log.Infof(log.Fields{}, "%v will be used", a[0].path)
			return a[0]
		}
	*/

	info := mountInfo{}
	index := 0

	for i := 0; i < len(a); i++ {
		info = a[(curMountIndex+i)%len(a)]
		if info.size < 600*1024*1024*1024 {
			continue
		}
		index = i
	}

	if 0 < len(a) {
		curMountIndex = (curMountIndex + index + 1) % len(a)
	}

	return info
}

// Mount 寻找合适的目录
func Mount() string {
	initMount()
	lock.Lock()
	defer lock.Unlock()
	return _mountInfos.mount().path
}

// InitMount find all mount info
func InitMount() error {
	tmps, _ := exec.Command("/usr/bin/find", "/mnt", "-name", "*.tmp").Output()
	tmpFiles := strings.Split(string(tmps), "\n")
	for _, tmp := range tmpFiles {
		log.Infof(log.Fields{}, "remove old tmps %v", tmp)
		exec.Command("/usr/bin/rm", "-rf", tmp).Run()
	}
	return initMount()
}

func initMount() error {
	mountPoints := make(map[string]mountInfo)
	filepath.Walk(mountRoot, func(absMountPath string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		ok, err := util.IsMountPoint(absMountPath)
		if err != nil {
			return nil
		}

		if ok {
			filepath.Walk(absMountPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && info != nil && !info.IsDir() {
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

		return nil
	})

	lock.Lock()
	_mountInfos = []mountInfo{}
	for _, v := range mountPoints {
		_mountInfos = append(_mountInfos, v)
	}
	lock.Unlock()

	return nil
}
