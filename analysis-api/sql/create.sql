CREATE SCHEMA IF NOT EXISTS "analysis";

GRANT
ALL
PRIVILEGES
ON
DATABASE
"analysis" TO harb;
GRANT ALL PRIVILEGES ON SCHEMA
"analysis" TO harb;

DO
$do$
BEGIN
        IF
NOT EXISTS (
                SELECT FROM pg_catalog.pg_user
                WHERE usename = ''harb'') THEN

            CREATE
USER harb WITH ENCRYPTED PASSWORD ''HueHue123'';
            GRANT ALL PRIVILEGES ON DATABASE
"analysis" TO harb;
            GRANT ALL PRIVILEGES ON SCHEMA
"analysis" TO harb;
END IF;
END
$do$;


CREATE TABLE IF NOT EXISTS "analysis"."currencies"
(
    "id"
    SERIAL
    PRIMARY
    KEY,
    "code"
    TEXT
    UNIQUE
    NOT
    NULL,
    "long_name"
    TEXT
    NOT
    NULL
);

CREATE TABLE IF NOT EXISTS "analysis"."symbols"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "currency_id"
    BIGINT
    NOT
    NULL,
    "isin"
    TEXT
    NOT
    NULL,
    "identifier"
    TEXT
    NOT
    NULL,
    "name"
    TEXT
    NOT
    NULL,
    "minimum_order_quantity"
    REAL
    NOT
    NULL,
    "market_name"
    TEXT
    NOT
    NULL,
    "market_hours_gmt"
    TEXT
    NOT
    NULL,
    "created_at"
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    "deleted_at" TIMESTAMPTZ NULL DEFAULT NULL,
    CONSTRAINT "fk_symbols_currency" FOREIGN KEY
(
    "currency_id"
) REFERENCES "analysis"."currencies"
(
    "id"
) ON UPDATE NO ACTION
  ON DELETE NO ACTION
    );

CREATE TABLE IF NOT EXISTS "analysis"."histories"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "symbol_id"
    BIGINT
    NULL
    DEFAULT
    NULL,
    "values"
    TEXT
    NOT
    NULL,
    "created_at"
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    CONSTRAINT "fk_histories_symbol" FOREIGN KEY
(
    "symbol_id"
) REFERENCES "analysis"."symbols"
(
    "id"
) ON UPDATE NO ACTION
  ON DELETE NO ACTION
    );

CREATE TABLE IF NOT EXISTS "analysis"."reports"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "symbol_id"
    BIGINT
    NOT
    NULL,
    "history_id"
    BIGINT
    NOT
    NULL,
    "exponential_moving_averages"
    TEXT
    NOT
    NULL,
    "simple_moving_averages"
    TEXT
    NOT
    NULL,
    "macd"
    TEXT
    NOT
    NULL,
    "rsi"
    TEXT
    NOT
    NULL,
    "created_at"
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    "deleted_at" TIMESTAMPTZ NULL DEFAULT NULL,
    CONSTRAINT "fk_report_symbol" FOREIGN KEY
(
    "symbol_id"
) REFERENCES "analysis"."symbols"
(
    "id"
) ON UPDATE NO ACTION
  ON DELETE NO ACTION,
    CONSTRAINT "fk_report_history" FOREIGN KEY
(
    "history_id"
) REFERENCES "analysis"."histories"
(
    "id"
)
  ON UPDATE NO ACTION
  ON DELETE NO ACTION
    );

CREATE TABLE IF NOT EXISTS "analysis"."strategies"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "name"
    TEXT
    NOT
    NULL,
    "created_at"
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    "deleted_at" TIMESTAMPTZ NULL DEFAULT NULL
    );

CREATE TABLE IF NOT EXISTS "analysis"."signals"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "symbol_id"
    BIGINT
    NOT
    NULL,
    "strategy_id"
    BIGINT
    NOT
    NULL,
    "type"
    BIGINT
    NOT
    NULL,
    "created_at"
    TIMESTAMPTZ
    NOT
    NULL
    DEFAULT
    now
(
),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT now
(
),
    "deleted_at" TIMESTAMPTZ NULL DEFAULT NULL,
    CONSTRAINT "fk_signal_symbol" FOREIGN KEY
(
    "symbol_id"
) REFERENCES "analysis"."symbols"
(
    "id"
) ON UPDATE NO ACTION
  ON DELETE NO ACTION,
    CONSTRAINT "fk_report_strategy" FOREIGN KEY
(
    "strategy_id"
) REFERENCES "analysis"."strategies"
(
    "id"
)
  ON UPDATE NO ACTION
  ON DELETE NO ACTION
    );

CREATE TABLE IF NOT EXISTS "analysis"."signal_reports"
(
    "id"
    BIGINT
    PRIMARY
    KEY,
    "signal_id"
    BIGINT
    NOT
    NULL,
    "report_id"
    BIGINT
    NOT
    NULL,

    CONSTRAINT
    "fk_signal_report_signal"
    FOREIGN
    KEY
(
    "signal_id"
) REFERENCES "analysis"."signals"
(
    "id"
) ON UPDATE NO ACTION
  ON DELETE NO ACTION,
    CONSTRAINT "fk_signal_report_report" FOREIGN KEY
(
    "report_id"
) REFERENCES "analysis"."reports"
(
    "id"
)
  ON UPDATE NO ACTION
  ON DELETE NO ACTION
    );