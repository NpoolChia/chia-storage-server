package chiaapi

import (
	"net/http"

	types "github.com/NpoolChia/chia-storage-server/types"

	"encoding/json"
	"fmt"

	log "github.com/EntropyPool/entropy-logger"
	httpdaemon "github.com/NpoolRD/http-daemon"
	"golang.org/x/xerrors"
)

func UploadChiaPlot(host string, input types.UploadPlotInput) (*types.UploadPlotOutput, error) {
	log.Infof(log.Fields{}, "req to http://%v%v", "", input.PlotURL)

	resp, err := httpdaemon.R().
		SetHeader("Content-Type", "application/json").
		SetBody(input).
		Post(fmt.Sprintf("http://%v%v", host, types.UploadPlotAPI))
	if err != nil {
		log.Errorf(log.Fields{}, "heartbeat error: %v", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, xerrors.Errorf("NON-200 return")
	}

	apiResp, err := httpdaemon.ParseResponse(resp)
	if err != nil {
		return nil, err
	}

	output := types.UploadPlotOutput{}
	b, _ := json.Marshal(apiResp.Body)
	err = json.Unmarshal(b, &output)

	return &output, err
}
