package db

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/dystopia-systems/alaskalog"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/entity"
	"io"
	"time"
)

type symbolRepository interface {
	GetBulkAsSymbolData() <-chan *entity.Symbol
	GetPaged(pageSize int, pageNumber int, orderBy string, asc bool) (*[]entity.Symbol, error)
	GetByISINAndName(isin, name string) (*entity.Symbol, error)
	InsertBulk(symbols []*entity.Symbol) (bool, error)
}

type SymbolRepository struct {
	symbolRepository
	db *pgx.ConnPool
}

func NewSymbolRepository(db *pgx.ConnPool) *SymbolRepository {
	return &SymbolRepository{
		db: db,
	}
}

func (r *SymbolRepository) GetBulkAsSymbolData() <-chan *entity.Symbol {
	result := make(chan *entity.Symbol)

	queryBuilder := squirrel.
		Select("*").
		From("symbols AS s").
		Join("currencies AS c on c.id = s.currency_id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		alaskalog.Logger.Warnf("failed generating query: %v", err)
		close(result)
		return result
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		alaskalog.Logger.Warnf("failed executing query: %v", err)
		close(result)
		return result
	}

	go func() {
		defer close(result)
		defer rows.Close()
		for rows.Next() {
			sym := entity.Symbol{Currency: entity.Currency{}}
			if err = rows.Scan(
				&sym.ID,
				&sym.CurrencyID,
				&sym.ISIN,
				&sym.Identifier,
				&sym.Name,
				&sym.MinimumOrderQuantity,
				&sym.MarketName,
				&sym.MarketHoursGMT,
				&sym.CreatedAt,
				&sym.UpdatedAt,
				&sym.DeletedAt,
				&sym.Currency.Code,
				&sym.Currency.LongName); err != nil && err != io.EOF {
				alaskalog.Logger.Warnf("failed scanning row: %v", err)
				return
			}
			sym.Currency.ID = sym.CurrencyID
			result <- &sym
		}
	}()

	return result
}

// GetPaged returns a paged response of symbols stored
func (r *SymbolRepository) GetPaged(pageSize int, pageNumber int, orderBy string, asc bool) (*[]entity.Symbol, error) {
	var e []entity.Symbol

	// generate query
	order := common.FormatOrderQuery(orderBy, asc)
	queryBuilder := squirrel.
		Select("*").
		From("symbols as s").
		OrderBy(order).
		Offset(uint64((pageNumber - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Join("currencies AS c ON c.id = s.currency_id").
		PlaceholderFormat(squirrel.Dollar)
	q, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	ctx, c := context.WithTimeout(context.Background(), time.Second)
	defer c()

	// send query
	rows, err := r.db.QueryEx(ctx, q, nil, args)
	if err != nil {
		return nil, err
	}
	println(rows.FieldDescriptions())
	defer rows.Close()

	return &e, nil
}

// GetByISINAndName gets a symbol by it's ISIN and name
func (r *SymbolRepository) GetByISINAndName(isin, name string) (*entity.Symbol, error) {
	var res entity.Symbol

	err := r.db.
		QueryRow(`SELECT * FROM symbols WHERE isin = $1 and name = $2 LIMIT 1`, isin, name).
		Scan(&res) // add all columns

	if err != nil {
		return nil, err
	}
	return &res, nil
}

// InsertBulk inserts the slice in a single transaction in batches and returns success and error
func (r *SymbolRepository) InsertBulk(symbols []*entity.Symbol) (bool, error) {
	timeoutContext, c := context.WithTimeout(context.Background(), 1*time.Second)
	defer c()

	tx, _ := r.db.BeginEx(timeoutContext, &pgx.TxOptions{})

	// truncate old symbol data
	_, err := tx.ExecEx(timeoutContext, "TRUNCATE symbols CASCADE", &pgx.QueryExOptions{})
	if err != nil {
		return false, err
	}

	// split inserts in batches
	workList := make(chan []*entity.Symbol)
	go func() {
		defer close(workList)
		batchSize := 1000
		var stack []*entity.Symbol
		for _, sym := range symbols {
			stack = append(stack, sym)
			if len(stack) == batchSize {
				workList <- stack
				stack = nil
			}
		}
		if len(stack) > 0 {
			workList <- stack
		}
	}()

	// generate query for insert from batches
	for list := range workList {
		queryForNew := squirrel.
			Insert("symbols").
			Columns("currency_id, isin, identifier, name, minimum_order_quantity, market_name, market_hours_gmt, created_at, updated_at, deleted_at").
			PlaceholderFormat(squirrel.Dollar)
		queryForOld := squirrel.
			Insert("symbols").
			Columns("id, currency_id, isin, identifier, name, minimum_order_quantity, market_name, market_hours_gmt, created_at, updated_at, deleted_at").
			PlaceholderFormat(squirrel.Dollar)

		for _, sym := range list {
			if sym.ID == 0 {
				queryForNew = queryForNew.Values(
					&sym.CurrencyID,
					&sym.ISIN,
					&sym.Identifier,
					&sym.Name,
					&sym.MinimumOrderQuantity,
					&sym.MarketName,
					&sym.MarketHoursGMT,
					&sym.CreatedAt,
					&sym.UpdatedAt,
					&sym.DeletedAt)
			} else {
				queryForOld = queryForOld.Values(
					&sym.ID,
					&sym.CurrencyID,
					&sym.ISIN,
					&sym.Identifier,
					&sym.Name,
					&sym.MinimumOrderQuantity,
					&sym.MarketName,
					&sym.MarketHoursGMT,
					&sym.CreatedAt,
					&sym.UpdatedAt,
					&sym.DeletedAt)
			}
		}

		// send query without ID column supplied
		query1, args1, _ := queryForNew.ToSql()
		if len(args1) > 0 {
			_, err = tx.ExecEx(timeoutContext, query1, &pgx.QueryExOptions{}, args1...)
			if err != nil {
				_ = tx.RollbackEx(timeoutContext)
				return false, err
			}
		}

		// send query with ID column supplied
		query2, args2, _ := queryForOld.ToSql()
		if len(args2) > 0 {
			_, err = tx.ExecEx(timeoutContext, query2, &pgx.QueryExOptions{}, args2...)
			if err != nil {
				_ = tx.RollbackEx(timeoutContext)
				return false, err
			}
		}
	}

	tx.CommitEx(timeoutContext)
	return true, nil
}
