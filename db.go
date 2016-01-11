package common

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type DBmysql struct {
	DSN  string
	Conn *sql.DB
}

type DBConf struct {
	Name   string `json:"name"`
	NGType string `json:"ng_type"`
	DBType string `json:"db_type"`
	DSN    string `json:"dsn"`
}

func InitDBPool(conf interface{}, pool interface{}) (err error) {
	var dbs []DBConf
	var byteConf []byte

	if byteConf, err = json.Marshal(conf); err != nil {
		return
	}
	if err = json.Unmarshal(byteConf, &dbs); err != nil {
		return
	}

	for _, db := range dbs {
		switch db.NGType {
		case "xorm":
			pool.(map[string]*xorm.Engine)[db.Name], err = xorm.NewEngine(db.DBType, db.DSN)
		case "mysql":
			DBConn := &DBmysql{DSN: db.DSN}
			DBConn.Conn, err = sql.Open(db.DBType, db.DSN)
			pool.(map[string]*DBmysql)[db.Name] = DBConn
		default:
			return fmt.Errorf("未知的ng_type: %v", db.NGType)
		}

		if err != nil {
			return
		}
	}

	return nil
}

func (this *DBmysql) Query(sqlStr string, args ...interface{}) (rst []map[string]string, err error) {
	var (
		stmt *sql.Stmt = nil
		rows *sql.Rows = nil
	)

	stmt, err = this.Conn.Prepare(sqlStr)
	if err != nil {
		return
	}
	defer stmt.Close()

	rows, err = stmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		return
	}

	cvals := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range cvals {
		scanArgs[i] = &cvals[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return
		}

		tmap := make(map[string]string, len(cols))
		for i, col := range cvals {
			if col == nil {
				tmap[cols[i]] = ""
			} else {
				tmap[cols[i]] = string(col)
			}
		}
		rst = append(rst, tmap)
	}

	return
}

func (this *DBmysql) Insert(sqlStr string, args ...interface{}) (int64, error) {
	stmt, err := this.Conn.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	rst, err := stmt.Exec(args...)
	if err != nil {
		return -1, err
	}

	return rst.LastInsertId()
}

func (this *DBmysql) Update(sqlStr string, args ...interface{}) (int64, error) {
	stmt, err := this.Conn.Prepare(sqlStr)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	rst, err := stmt.Exec(args...)
	if err != nil {
		return -1, err
	}

	return rst.RowsAffected()
}
