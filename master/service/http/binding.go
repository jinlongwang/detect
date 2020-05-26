package http

type StrategyJson struct {
	Name     string                 `json:"name"`
	Note     string                 `json:"note"`
	Mode     int64                  `json:"mode"`
	IsDelete bool                   `json:"is_delete"`
	Context  map[string]interface{} `json:"context"`
}
