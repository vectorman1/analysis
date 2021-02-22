
-- Symbols that share their identifier - e.g. AAPL with more than 2 symbols
SELECT s.*
FROM
    (select identifier
     FROM analysis.symbols
     GROUP BY identifier
     HAVING COUNT(*) > 2
    ) AS i
        JOIN analysis.symbols as  s
             on s.identifier = i.identifier order by identifier;

-- Symbols their isin with more than 2 other symbols
SELECT s.*
FROM
    (select isin
     FROM analysis.symbols
     GROUP BY isin
     HAVING COUNT(*) > 2
    ) AS i
        JOIN analysis.symbols as  s
             on s.isin = i.isin order by isin;

