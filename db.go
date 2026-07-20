package common

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DBmysql struct {
	DSN  string
	Conn *sql.DB
}

type DBConf struct {
	DBType string `json:"db_type"`
	DSN    string `json:"dsn"`
}

func InitMysql(conf DBConf) (db *DBmysql, err error) {
	conn, err := sql.Open(conf.DBType, conf.DSN)
	db = &DBmysql{DSN: conf.DSN, Conn: conn}
	return db, err
}

func (this *DBmysql) Query(sqlStr string, args ...interface{}) (rst []map[string]string, err error) {
	var (
		stmt *sql.Stmt
		rows *sql.Rows
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
