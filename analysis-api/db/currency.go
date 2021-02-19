package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/entity"
)

type currencyRepository interface {
	GetByCode(code string, c *pgx.ConnPool) (*entity.Currency, error)
	Create(curr *entity.Currency, c *pgx.ConnPool) (uint, error)
}

type CurrencyRepository struct {
	currencyRepository
	db *pgx.ConnPool
}

func NewCurrencyRepository(pgDb *pgx.ConnPool) *CurrencyRepository {
	return &CurrencyRepository{
		db: pgDb,
	}
}

func (r *CurrencyRepository) GetByCode(code string) (*entity.Currency, error) {
	queryBuilder := squirrel.
		Select("id, code, long_name").
		From("currencies").
		Where(squirrel.Eq{"code": code}).
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows := r.db.QueryRow(query, args...)

	var res entity.Currency
	err = rows.Scan(&res.ID, &res.Code, &res.LongName)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *CurrencyRepository) Create(curr *entity.Currency) (uint, error) {
	queryBuilder := squirrel.
		Insert("currencies").
		Columns("code, long_name").
		Values(curr.Code, curr.LongName).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	id := uint(0)
	err = r.db.QueryRow(query+" RETURNING id;", args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *CurrencyRepository) BeginTransaction() (*pgx.Tx, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *CurrencyRepository) RollbackTransaction(tx *pgx.Tx) error {
	err := tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func (r *CurrencyRepository) CommitTransaction(tx *pgx.Tx) error {
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
