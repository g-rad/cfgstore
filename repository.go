package cfgstore

import (
	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/g-rad/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(cs string) (*Repository, error) {

	var err error
	r := Repository{}
	r.db, err = sqlx.Connect("sqlserver", cs)

	if err == nil {

		// as per https://github.com/denisenkom/go-mssqldb/issues/167
		// When / if this is not set, the library considers the connection to be reusable for an indefinite period,
		// a problem since Azure's SQL DB closes an idle connection after 30 mins (by design).
		r.db.SetConnMaxLifetime(20 * time.Minute)
	}

	return &r, err
}

func (r Repository) ConfigGet(key string) ([]*ConfigKeyValue, error) {

	var dest []*ConfigKeyValue

	if err := r.db.Select(&dest, "EXEC sp_GetSettings @EnvKey", sql.Named("EnvKey", key)); err != nil {
		return nil, errors.Wrap(err, "error requesting db")
	}

	return dest, nil
}
