package mount

import (
	"os"
	"os/exec"
	"path/filepath"
	_ "sort"
	"strings"
	"sync"

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
		size uint64
	}

	// all mount point info
	mountInfos []mountInfo
)

var (
	_mountInfos   mountInfos
	lock          sync.Mutex
	curMountIndex int
	reservedSpace uint64
)

func (a mountInfos) Len() int      { return len(a) }
func (a mountInfos) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a mountInfos) Less(i, j int) bool {
	// first sort by size, then tmp file count
	return a[i].size > a[j].size
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
		myInfo := a[(curMountIndex+i)%len(a)]
		if myInfo.size < reservedSpace {
			log.Infof(log.Fields{}, "%v available %v < 600G", myInfo.path, myInfo.size)
			continue
		}
		info = myInfo
		index = i
		break
	}

	curMountIndex = (curMountIndex + index + 1) % len(a)

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
func InitMount(reserved uint64) error {
	reservedSpace = reserved
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
			avail, err := util.New(absMountPath).GetAvail()
			if err != nil {
				return nil
			}

			mountPoints[absMountPath] = mountInfo{
				path: absMountPath,
				size: avail,
			}
		}

		return nil
	})

	lock.Lock()
	_mountInfos = []mountInfo{}
	for _, v := range mountPoints {
		log.Infof(log.Fields{}, "append valid mountpoint %v | %v", v.path, v.size)
		_mountInfos = append(_mountInfos, v)
	}
	lock.Unlock()

	return nil
}
