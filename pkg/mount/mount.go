package mount

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	_ "sort"
	"strings"
	"sync"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolChia/chia-storage-server/util"
)

const (
	tmpFileSize = 101 * 1024 * 1024 * util.KB // 101G->kb
	mountRoot   = "/mnt"
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
	return a[i].path > a[j].path
}

// Choose the right moint point
func (a mountInfos) mount() mountInfo {
	info := mountInfo{}
	index := 0

	for i := 0; i < len(a); i++ {
		myInfo := a[(curMountIndex+i)%len(a)]
		if myInfo.size < reservedSpace {
			log.Infof(log.Fields{}, "%v available %v < %v", myInfo.path, myInfo.size, reservedSpace)
			continue
		}
		info = myInfo
		index = i
		break
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
	path := _mountInfos.mount().path
	log.Infof(log.Fields{}, "select mount path %v", path)
	lock.Unlock()
	return path
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

		// not a dir
		if !info.IsDir() {
			return nil
		}

		ok, err := util.IsMountPoint(absMountPath)
		if err != nil {
			log.Infof(log.Fields{}, "check %v is mount point error %v", absMountPath, err)
			return nil
		}

		if ok {
			avail, err := util.New(absMountPath).GetAvail()
			if err != nil {
				log.Infof(log.Fields{}, "get mount point %v avail space error %v", absMountPath, err)
				return nil
			}

			// 读取目录下的 tmp 文件个数, 每个累计101G
			mountPoints[absMountPath] = mountInfo{
				path: absMountPath,
				size: avail - getTmpFileSize(absMountPath),
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

	// sort by mount point path
	sort.Sort(_mountInfos)
	lock.Unlock()

	return nil
}

// getTmpFileSize 获取指定目录下的 tmp 文件个数
func getTmpFileSize(root string) (size uint64) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == TmpFileExt {
			size += tmpFileSize
		}
		return nil
	})
	return
}
