package database
 
import (
	"context"
	"database/sql"
	_"github.com/jackc/pgx/v5/stdlib"
)

func Connect(ctx context.Context, DatabaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}