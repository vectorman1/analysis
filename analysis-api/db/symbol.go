package db

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/generated/symbol_service"
	"github.com/vectorman1/analysis/analysis-api/model"
)

type symbolRepository interface {
	GetPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*[]model.Symbol, error)
	GetByISINAndName(isin, name string) (*model.Symbol, error)
	InsertBulk(symbols []*model.Symbol) (bool, error)
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

// GetPaged returns a paged response of symbols stored
func (r *SymbolRepository) GetPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*[]model.Symbol, error) {
	// generate query
	order := common.FormatOrderQuery(req.Filter.Order, req.Filter.Ascending)
	queryBuilder := squirrel.
		Select("*").
		From("analysis.symbols as s").
		OrderBy(order).
		Offset((req.Filter.PageNumber - 1) * req.Filter.PageSize).
		Limit(req.Filter.PageSize).
		Join("analysis.currencies AS c ON c.id = s.currency_id").
		PlaceholderFormat(squirrel.Dollar)
	q, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	tctx, c := context.WithTimeout(ctx, time.Second)
	defer c()

	conn, err := r.db.Acquire()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryEx(tctx, q, nil, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Symbol
	for rows.Next() {
		sym := model.Symbol{Currency: model.Currency{}}
		if err = rows.Scan(
			&sym.ID,
			&sym.CurrencyID,
			&sym.Isin,
			&sym.Identifier,
			&sym.Name,
			&sym.MinimumOrderQuantity,
			&sym.MarketName,
			&sym.MarketHoursGmt,
			&sym.CreatedAt,
			&sym.UpdatedAt,
			&sym.DeletedAt,
			&sym.Currency.ID,
			&sym.Currency.Code,
			&sym.Currency.LongName); err != nil {
			return nil, err
		}
		result = append(result, sym)
	}

	return &result, nil
}

// InsertBulk inserts the slice in a single transaction in batches and returns success and error
func (r *SymbolRepository) InsertBulk(ctx context.Context, symbols []*model.Symbol) (bool, error) {
	timeoutContext, c := context.WithTimeout(ctx, 1*time.Second)
	defer c()

	conn, err := r.db.Acquire()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	tx, _ := conn.BeginEx(timeoutContext, &pgx.TxOptions{})

	// split inserts in batches
	workList := make(chan []*model.Symbol)
	go func() {
		defer close(workList)
		batchSize := 1000
		var stack []*model.Symbol
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
		q := squirrel.
			Insert("analysis.symbols").
			Columns("currency_id, isin, identifier, name, minimum_order_quantity, market_name, market_hours_gmt, created_at, updated_at, deleted_at").
			PlaceholderFormat(squirrel.Dollar)
		for _, sym := range list {
			q = q.Values(
				&sym.CurrencyID,
				&sym.Isin,
				&sym.Identifier,
				&sym.Name,
				&sym.MinimumOrderQuantity,
				&sym.MarketName,
				&sym.MarketHoursGmt,
				&sym.CreatedAt,
				&sym.UpdatedAt,
				&sym.DeletedAt)
		}

		query, args, _ := q.ToSql()
		if len(args) > 0 {
			_, err = tx.ExecEx(timeoutContext, query, &pgx.QueryExOptions{}, args...)
			if err != nil {
				_ = tx.RollbackEx(timeoutContext)
				return false, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}
