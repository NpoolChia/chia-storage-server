package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/EntropyPool/entropy-logger"
	chiastorageProxyTypes "github.com/NpoolChia/chia-storage-proxy/types"
	"github.com/NpoolChia/chia-storage-server/pkg/mount"
	types "github.com/NpoolChia/chia-storage-server/types"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"github.com/go-resty/resty/v2"
)

type ChiaStorageServerConfig struct {
	Port int `json:"port"`
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

	return server
}

var (
	errPlotURLEmpty = errors.New("plot url is empty")
)

func (s *ChiaStorageServer) UploadPlotRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	req.URL.Query().Get("")
	// get chia plot file
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	input := types.UploadPlotInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -2
	}
	if input.PlotURL == "" ||
		input.FinishURL == "" ||
		input.FailURL == "" {
		return nil, errPlotURLEmpty.Error(), -3
	}

	plotFile := filepath.Base(input.PlotURL)

	go func(input types.UploadPlotInput) {
		var (
			err  error
			resp *resty.Response
		)
		defer func() {
			// notify client write plot file result
			var (
				notifyURL = ""
				body      = make([]byte, 0)
			)
			if err != nil {
				notifyURL = input.FailURL
				fail := chiastorageProxyTypes.FailPlotInput{
					PlotFile: input.PlotURL,
				}
				body, _ = json.Marshal(fail)
			} else {
				notifyURL = input.FinishURL
				finish := chiastorageProxyTypes.FinishPlotInput{
					PlotFile: input.PlotURL,
				}
				body, _ = json.Marshal(finish)
			}

			_, err = httpdaemon.R().
				SetHeader("Content-Type", "application/json").
				SetBody(body).
				Post(notifyURL)
			if err != nil {
				return
			}
		}()

		// 选择存放的目录
		path := mount.Mount()
		// 没有挂载的盘符
		if path == "" {
			// TODO
			return
		}

		tmp := filepath.Join(path, temp(plotFile))
		plot, err := os.Create(tmp)
		if err != nil {
			return
		}

		defer plot.Close()
		resp, err = httpdaemon.R().Get(input.PlotURL)
		if err != nil {
			return
		}
		defer resp.RawBody().Close()
		if _, err = io.Copy(plot, resp.RawBody()); err != nil {
			return
		}

		// 移除临时文件
		defer os.Remove(tmp)
		if err = os.Rename(tmp, plotFile); err != nil {
			return
		}
	}(input)
	return nil, "", 0
}

func temp(src string) string {
	return src + ".tmp"
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
