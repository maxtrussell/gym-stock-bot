package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/maxtrussell/gym-stock-bot/models/item"
)

type StockRow struct {
	ProductName string
	ItemName    string
	Price       string
	InStock     bool
	Timestamp   time.Time
}

func (r StockRow) ID() string {
	return r.ProductName + ": " + r.ItemName
}

func Setup() *sql.DB {
	db := connect("db.sqlite")
	createTable(db)
	return db
}

func UpdateStock(db *sql.DB, items []item.Item) {
	rows := queryLatestStock(db)
	for _, i := range items {
		r, ok := rows[i.ID()]
		if ok && i.IsAvailable() != r.InStock {
			// Insert row when availability is mismatched
			InsertStockRow(db, i)
		} else if !ok {
			// Insert new items, not yet in db
			InsertStockRow(db, i)
		}
	}
}

func InsertStockRow(db *sql.DB, i item.Item) {
	q := `
    INSERT INTO stock(
        ProductName,
        ItemName,
        Price,
        InStock
    ) values (?, ?, ?, ?);`
	stmt, err := db.Prepare(q)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Product.Name, i.Name, i.Price, i.IsAvailable())
	if err != nil {
		log.Fatal(err)
	}
}

func queryLatestStock(db *sql.DB) map[string]*StockRow {
	q := `
    SELECT ProductName, ItemName, Price, InStock, DATETIME(Timestamp, 'localtime')
    FROM stock
    ORDER BY Timestamp DESC;`

	rows := queryStock(db, q)
	m := map[string]*StockRow{}
	for _, r := range rows {
		if _, ok := m[r.ID()]; !ok {
			m[r.ID()] = r
		}
	}
	return m
}

func connect(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

func queryStock(db *sql.DB, q string) []*StockRow {
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var stock_rows []*StockRow
	for rows.Next() {
		stock_row := StockRow{}
		err = rows.Scan(
			&stock_row.ProductName,
			&stock_row.ItemName,
			&stock_row.Price,
			&stock_row.InStock,
			&stock_row.Timestamp,
		)
		if err != nil {
			log.Fatal(err)
		}
		stock_rows = append(stock_rows, &stock_row)
	}
	return stock_rows
}

func createTable(db *sql.DB) {
	sql_table := `
    CREATE TABLE IF NOT EXISTS stock(
        ID INTEGER PRIMARY KEY AUTOINCREMENT,
        ProductName TEXT NOT NULL,
        ItemName TEXT NOT NULL,
        Price TEXT,
        InStock INTEGER NOT NULL,
        Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := db.Exec(sql_table); err != nil {
		log.Fatal(err)
	}
}
