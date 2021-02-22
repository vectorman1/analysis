package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/pgtype"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/generated/symbol_service"
	"github.com/vectorman1/analysis/analysis-api/model"
)

type symbolRepository interface {
	GetPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*[]model.Symbol, error)
	GetByISINAndName(isin, name string) (*model.Symbol, error)
	InsertBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error)
	DeleteBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error)
	UpdateBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error)

	BeginTx(ctx *context.Context, options *pgx.TxOptions) (*pgx.Tx, error)
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

	// limit query time to a second
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

	// read all resulting rows
	var result []model.Symbol
	for rows.Next() {
		sym := model.Symbol{Currency: model.Currency{}}
		if err = rows.Scan(
			&sym.ID,
			&sym.Uuid,
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
func (r *SymbolRepository) InsertBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error) {
	conn, err := r.db.Acquire()
	if err != nil {
		return false, err
	}
	defer conn.Close()

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
			Columns("uuid, currency_id, isin, identifier, name, minimum_order_quantity, market_name, market_hours_gmt, created_at, updated_at, deleted_at").
			PlaceholderFormat(squirrel.Dollar)
		for _, sym := range list {
			q = q.Values(
				&sym.Uuid,
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
			_, err = tx.ExecEx(*timeoutContext, query, &pgx.QueryExOptions{}, args...)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// DeleteBulk sets the Deleted At values for bulk symbols to now
func (r *SymbolRepository) DeleteBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error) {
	// split updates in batches
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

	deletedAt := pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
	// generate query for insert from batches
	for list := range workList {
		q := squirrel.Update("analysis.symbols")
		for _, sym := range list {
			q = q.
				Set("deleted_at", deletedAt).
				PlaceholderFormat(squirrel.Dollar).
				Where(squirrel.Eq{"uuid": sym.Uuid})
		}

		query, args, _ := q.ToSql()
		if len(args) > 0 {
			_, err := tx.ExecEx(*timeoutContext, query, &pgx.QueryExOptions{}, args...)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// UpdateBulk updates all columns of the symbol with the matching uuid
// with the passed symbol values
func (r *SymbolRepository) UpdateBulk(tx *pgx.Tx, timeoutContext *context.Context, symbols []*model.Symbol) (bool, error) {
	// split updates in batches
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

	for list := range workList {
		for _, sym := range list {
			var u string
			sym.Uuid.AssignTo(&u)

			q := squirrel.
				Update("analysis.symbols").
				PlaceholderFormat(squirrel.Dollar)

			q = q.
				Set("currency_id", sym.CurrencyID).
				Set("name", sym.Name).
				Set("minimum_order_quantity", sym.MinimumOrderQuantity.Float).
				Set("market_name", sym.MarketName).
				Set("market_hours_gmt", sym.MarketHoursGmt).
				Set("updated_at", time.Now()).
				Where(squirrel.Eq{"uuid::text": u})

			query, args, _ := q.ToSql()
			if len(args) > 0 {
				_, err := tx.ExecEx(*timeoutContext, query, &pgx.QueryExOptions{}, args...)
				if err != nil {
					return false, err
				}
			}
		}
	}

	return true, nil
}

func (r *SymbolRepository) BeginTx(ctx *context.Context, options *pgx.TxOptions) (*pgx.Tx, error) {
	tx, err := r.db.BeginEx(*ctx, options)
	if err != nil {
		return nil, err
	}

	return tx, err
}
