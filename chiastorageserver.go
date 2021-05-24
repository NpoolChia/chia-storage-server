package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolChia/chia-storage-server/pkg/mount"
	"github.com/NpoolChia/chia-storage-server/tasks"
	types "github.com/NpoolChia/chia-storage-server/types"
	"github.com/NpoolChia/chia-storage-server/util"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"github.com/boltdb/bolt"
)

type ChiaStorageServerConfig struct {
	Port int `json:"port"`
	// 数据库地址
	DBPath        string `json:"db_path"`
	ClusterName   string `json:"cluster_name"`
	ReservedSpace uint64 `json:"reserved_space"`
}

type ChiaStorageServer struct {
	config ChiaStorageServerConfig
}

func NewChiaStorageServer(configFile string) *ChiaStorageServer {
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read file %v: %v", configFile, err)
		return nil
	}

	config := ChiaStorageServerConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse file %v: %v", configFile, err)
		return nil
	}

	server := &ChiaStorageServer{
		config: config,
	}

	log.Infof(log.Fields{}, "successful to create chia storage server")
	mount.InitMount(config.ReservedSpace)

	return server
}

var (
	errPlotURLEmpty = errors.New("plot url is empty")
)

func (s *ChiaStorageServer) UploadPlotRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	// get chia plot file
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to read body from %v", req.URL)
		return nil, err.Error(), -1
	}

	input := types.UploadPlotInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.Errorf(log.Fields{}, "fail to parse body from %v", req.URL)
		return nil, err.Error(), -2
	}
	if input.PlotURL == "" ||
		input.FinishURL == "" ||
		input.FailURL == "" {
		log.Errorf(log.Fields{}, "invalid input parameters from %v", req.URL)
		return nil, errPlotURLEmpty.Error(), -3
	}

	// 入库，调度队列处理
	db, err := util.BoltClient()
	if err != nil {
		return nil, err.Error(), -3
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket(util.DefaultBucket)
		if r := bk.Get([]byte(input.PlotURL)); r != nil {
			return fmt.Errorf("chia plot file url: %s already added", input.PlotURL)
		}
		meta := tasks.Meta{
			Status:      tasks.TaskTodo,
			ClusterName: s.config.ClusterName,
			PlotURL:     input.PlotURL,
			FinishURL:   input.FinishURL,
			FailURL:     input.FailURL,
		}
		ms, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		return bk.Put([]byte(input.PlotURL), ms)
	}); err != nil {
		return nil, err.Error(), -4
	}

	return nil, "", 0
}

func (s *ChiaStorageServer) Run() error {
	// 获取 chia plot file
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.UploadPlotAPI,
		Handler:  s.UploadPlotRequest,
		Method:   http.MethodPost,
	})

	httpdaemon.Run(s.config.Port)
	return nil
}
