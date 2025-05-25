package entity

import "database/sql"

type Sample struct {
	Index      int            `db:"INDEX"`
	String     sql.NullString `db:"STRING"`
	Int        sql.NullInt32  `db:"INT"`
	DateColumn sql.NullTime   `db:"DATE_COLUMN"`
}
