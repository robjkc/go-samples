package models

import (
	"database/sql"
	"strings"

	"github./securecom/haproxy-audit/db"
)

type User struct {
	ID                  int           `db:"id"`
	AccessibleId        sql.NullInt32 `db:"accessible_id"`
	Role                string        `db:"role"`
	AccessibleType      string        `db:"accessible_type"`
	AuthenticationToken string        `db:"authentication_token"`
	PersonId            sql.NullInt32 `db:"person_id"`
	DealerTempExpires   db.NullTime   `db:"dealer_temp_expires"`
}

func GetUser(con *db.DbConnection, authToken string) (*User, error) {

	users := []User{}
	err := con.ExecuteSelect(&users, `select id,
		accessible_id,
		role,
		accessible_type,
		authentication_token,
		person_id,
		dealer_temp_expires
		from users where authentication_token = :authToken`, db.Args{"authToken": authToken})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
}

func (u *User) IsCustomer() bool {
	return strings.EqualFold(u.AccessibleType, "customer")
}

func (u *User) IsDealerTemp() bool {
	return u.DealerTempExpires.Valid
}
