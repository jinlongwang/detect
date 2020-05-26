package model

import "github.com/go-xorm/xorm"

func InitSqlEngine(connString string, driverName string) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine(driverName, connString)
	if err != nil {
		return nil, err
	}
	return engine, nil
}
