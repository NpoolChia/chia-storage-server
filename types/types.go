package types

type UploadPlotInput struct {
	PlotURL   string `json:"plot_url"`
	FinishURL string `json:"finish_url"`
	FailURL   string `json:"fail_url"`
}

type UploadPlotOutput struct {
}
