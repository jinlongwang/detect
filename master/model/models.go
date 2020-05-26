package model

type Strategy struct {
	Id       int64
	Name     string
	Note     string
	Mode     int64
	IsDelete bool   `xorm:"int 'is_delete'"`
	Context  string `xorm:"text"`
	Interval int64
}

type Metrics struct {
	Id         int64
	StrategyId int64 `xorm:"strategy_id"`
	Metric     string
	Value      float64
	Step       int64
	MType      string `xorm:"type"`
	Timestamp  int64
	Tags       string
}
