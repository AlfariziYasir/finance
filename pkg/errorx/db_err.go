package errorx

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func DbError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return NewError(ErrTypeNotFound, "resource not found in database", err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return NewError(ErrTypeConflict, fmt.Sprintf("duplicate data: %s", pgErr.Detail), pgErr)
		case "23503":
			return NewError(ErrTypeNotFound, "referenced resource not found (foreign key constraint)", pgErr)
		}
	}

	return NewError(ErrTypeInternal, "internal database error", err)
}
