package common

import (
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type DBConf struct {
	Name   string `json:"name"`
	NGType string `json:"ng_type"`
	DBType string `json:"db_type"`
	DSN    string `json:"dsn"`
}

func InitDBPool(conf interface{}, pool interface{} /* map[string]*xorm.Engine */) (err error) {
	var dbs []DBConf
	var byteConf []byte

	if byteConf, err = json.Marshal(conf); err != nil {
		return
	}
	if err = json.Unmarshal(byteConf, &dbs); err != nil {
		return
	}

	for _, db := range dbs {
		if db.NGType == "xorm" {
			pool.(map[string]*xorm.Engine)[db.Name], err = xorm.NewEngine(db.DBType, db.DSN)
		}

		if err != nil {
			return
		}
	}

	return nil
}
