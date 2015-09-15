package common

import (
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var (
	Xorms = make(map[string]*xorm.Engine)
)

func InitDbPool(conf interface{}) (err error) {
	var dbs []struct {
		Name   string `json:"name"`
		NGType string `json:"ng_type"`
		DBType string `json:"db_type"`
		DSN    string `json:"dsn"`
	}

	var byteConf []byte
	if byteConf, err = json.Marshal(conf); err != nil {
		return
	}

	if err = json.Unmarshal(byteConf, &dbs); err != nil {
		return
	}

	for _, db := range dbs {
		if db.NGType == "xorm" {
			Xorms[db.Name], err = xorm.NewEngine(db.DBType, db.DSN)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
