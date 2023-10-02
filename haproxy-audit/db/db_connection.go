package db

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github./securecom/haproxy-audit/config"
)

type DbConnection struct {
	// https://www.alexedwards.net/blog/organising-database-access
	sqlDb *sqlx.DB
}

type Args map[string]interface{}

var (
	DbConn *DbConnection
)

func ConnectDb(url string) (*DbConnection, error) {
	db, err := sqlx.Connect("sqlserver", url)
	//models.DB, err = sqlx.Connect("sqlserver", "Data Source=172.31.55.57;Initial Catalog=farad_production;Integrated Security=False;Pooling=True;MultipleActiveResultSets=True;User ID=odata_api;Password=thn%$nI4n@cd2;App=ODataAPI")
	if err != nil {
		return nil, err
	}

	// https://github.com/jmoiron/sqlx/issues/300
	db.SetMaxOpenConns(10)   // The default is 0 (unlimited)
	db.SetMaxIdleConns(5)    // defaultMaxIdleConns = 2
	db.SetConnMaxLifetime(0) // 0, connections are reused forever.

	DbConn = &DbConnection{sqlDb: db}
	return DbConn, nil
}

func (c *DbConnection) Connect(config *config.Config) (err error) {

	c.sqlDb, err = sqlx.Connect("sqlserver", config.ConnectionString)
	//models.DB, err = sqlx.Connect("sqlserver", "Data Source=172.31.55.57;Initial Catalog=farad_production;Integrated Security=False;Pooling=True;MultipleActiveResultSets=True;User ID=odata_api;Password=thn%$nI4n@cd2;App=ODataAPI")
	if err != nil {
		return err
	}

	// https://github.com/jmoiron/sqlx/issues/300
	c.sqlDb.SetMaxOpenConns(10)   // The default is 0 (unlimited)
	c.sqlDb.SetMaxIdleConns(5)    // defaultMaxIdleConns = 2
	c.sqlDb.SetConnMaxLifetime(0) // 0, connections are reused forever.

	return nil
}

func (c *DbConnection) ExecuteSelect(dest interface{}, query string, args Args) error {

	stmt, err := c.sqlDb.PrepareNamed(query)

	if err != nil {
		return errors.Wrapf(err, "error on prepare")
	}

	err = stmt.Select(dest, args)
	if err != nil {
		return errors.Wrapf(err, "error on get")
	}
	return nil
}

func (c *DbConnection) ExecuteSelectWithArgs(dest interface{}, query string, keyvals ...interface{}) error {

	// https://github.com/golang/go/issues/26459
	args := make(map[string]interface{}, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]
		args[fmt.Sprint(k)] = v
	}

	stmt, err := c.sqlDb.PrepareNamed(query)

	if err != nil {
		return errors.Wrapf(err, "error on prepare")
	}

	err = stmt.Select(dest, args)
	if err != nil {
		return errors.Wrapf(err, "error on get")
	}
	return nil
}

func (c *DbConnection) ExecuteUpdate(query string, args Args) error {
	stmt, err := c.sqlDb.PrepareNamed(query)

	if err != nil {
		return errors.Wrapf(err, "error on prepare")
	}

	_, err = stmt.Exec(args)
	if err != nil {
		return errors.Wrapf(err, "error on exec")
	}
	return nil
}

func (c *DbConnection) Close() {
	c.sqlDb.Close()
}
