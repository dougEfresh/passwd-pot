// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package potdb

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DB wrapper for psql/mysql support
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Ping() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Insert(query string, args ...interface{}) (sql.Result, error)
	Get() *sql.DB
}

type potDB struct {
	db    *sql.DB
	mysql bool
}

type psqlDialect struct {
	id int64
}

func (psql psqlDialect) LastInsertId() (int64, error) {
	return psql.id, nil
}

func (psql psqlDialect) RowsAffected() (int64, error) {
	return -1, nil
}

// Open a new DB connection pool
func Open(dsn string) (DB, error) {
	db, err := loadDSN(dsn)
	p := &potDB{
		db:    db,
		mysql: !strings.Contains(dsn, "postgres"),
	}

	return p, err
}

func (p *potDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(p.replaceParams(query), args...)
}

func (p *potDB) Ping() error {
	return p.db.Ping()
}

func (p *potDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(p.replaceParams(query), args...)
}

func (p *potDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(p.replaceParams(query), args...)
}

func (p *potDB) Insert(query string, args ...interface{}) (sql.Result, error) {
	if p.mysql {
		return p.db.Exec(query, args...)
	}
	psqlQuery := p.replaceParams(query) + " RETURNING ID"
	r, err := p.db.Query(psqlQuery, args...)
	var id int64
	if err != nil {
		return psqlDialect{0}, err
	}
	defer r.Close()
	if !r.Next() {
		return psqlDialect{0}, fmt.Errorf("Failed inserting %s", psqlQuery)
	}
	err = r.Scan(&id)
	return psqlDialect{id: id}, err
}

func (p *potDB) Get() *sql.DB {
	return p.db
}

func (p *potDB) replaceParams(sql string) string {
	if p.mysql {
		return sql
	}
	pos := 1
	for strings.Contains(sql, "?") {
		sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", pos), 1)
		pos++
	}
	return sql
}

func loadDSN(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if strings.Contains(dsn, "postgres") {
		db, err = sql.Open("postgres", dsn)
	} else {
		db, err = sql.Open("mysql", dsn)
	}

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}
