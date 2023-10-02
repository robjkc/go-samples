package models

import (
	"database/sql"

	"github./securecom/haproxy-audit/db"
)

type Person struct {
	ID           int            `db:"id"`
	EmailAddress sql.NullString `db:"email_address"`
	FirstName    sql.NullString `db:"first_name"`
	LastName     sql.NullString `db:"last_name"`
}

func GetPerson(con *db.DbConnection, personId int) (*Person, error) {

	persons := []Person{}
	err := con.ExecuteSelect(&persons, `select id,
		email_address,
		first_name,
		last_name
		from persons where id = :personId`, db.Args{"personId": personId})
	if err != nil {
		return nil, err
	}

	if len(persons) == 0 {
		return nil, nil
	}
	return &persons[0], nil
}
