package db

import (
	"github.com/Masterminds/squirrel"
	"github.com/dystopia-systems/alaskalog"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/model"
)

type currencyRepository interface {
	GetByCode(code string) (*model.Currency, error)
	GetOrCreate(code string) (*model.Currency, error)
	Create(curr *model.Currency) (uint, error)
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

func (r *CurrencyRepository) GetByCode(code string) (*model.Currency, error) {
	queryBuilder := squirrel.
		Select("id, code, long_name").
		From("analysis.currencies").
		Where(squirrel.Eq{"code": code}).
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows := r.db.QueryRow(query, args...)

	var res model.Currency
	err = rows.Scan(&res.ID, &res.Code, &res.LongName)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *CurrencyRepository) GetOrCreate(code string) (*model.Currency, error) {
	curr, err := r.GetByCode(code)
	if err != nil {
		newCurr := &model.Currency{}
		newCurr.Code = code
		newCurr.LongName = "temp name"

		id, createErr := r.Create(newCurr)
		if createErr != nil {
			alaskalog.Logger.Warnf("failed creating currency: %v", err)
			return nil, createErr
		}

		newCurr.ID = id
		return newCurr, nil
	} else {
		return curr, nil
	}
}

func (r *CurrencyRepository) Create(curr *model.Currency) (uint, error) {
	queryBuilder := squirrel.
		Insert("analysis.currencies").
		Columns("code, long_name").
		Values(curr.Code, curr.LongName).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	conn, err := r.db.Acquire()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	id := uint(0)
	err = conn.QueryRow(query+" RETURNING id;", args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
