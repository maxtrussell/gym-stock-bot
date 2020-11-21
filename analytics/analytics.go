package analytics

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/maxtrussell/gym-stock-bot/database"
)

func ItemReport(db *sql.DB, id string) {
	rows := database.QueryItemByID(db, id)
	since := rows[len(rows)-1].Timestamp

	last_in_stock := "never"
	last_out_of_stock := "never"

	times_in_stock := 0
	times_out_of_stock := 0

	time_in_stock := 0.0
	time_out_of_stock := 0.0

	for i, r := range rows {
		// 1. update last in/out of stock
		if r.InStock && last_in_stock == "never" {
			last_in_stock = r.Timestamp
		} else if !r.InStock && last_out_of_stock == "never" {
			last_out_of_stock = r.Timestamp
		}

		// 2. update times in/out of stock
		if r.InStock {
			times_in_stock++
		} else {
			times_out_of_stock++
		}

		// 3. update time in/out of stock
		if i != 0 {
			timeSince := parseTime(r.Timestamp).Sub(parseTime(rows[len(rows)-1].Timestamp))
			if r.InStock {
				time_out_of_stock += timeSince.Hours()
			} else {
				time_in_stock += timeSince.Hours()
			}
		}
	}

	fmt.Printf("Printing report for %s\n", id)
	fmt.Printf("Data since: %s\n", since)

	// 1. Last in stock/Last out of stock
	fmt.Printf("Last in stock: %s\n", last_in_stock)
	fmt.Printf("Last out of stock: %s\n", last_out_of_stock)

	// 2. Number of times in stock
	fmt.Printf("Times in stock: %d\n", times_in_stock)
	fmt.Printf("Times out of stock: %d\n", times_out_of_stock)

	// 3. Average days in/out of stock
	fmt.Printf("Time in stock: %f\n", time_in_stock)
	fmt.Printf("Time out of stock: %f\n", time_out_of_stock)
	fmt.Printf("Avg time in stock: %f\n", time_in_stock/float64(times_in_stock))
	fmt.Printf("Avg time out of stock: %f\n", time_out_of_stock/float64(times_out_of_stock))

	// 4. Predicted next in stock
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		log.Fatal(err)
	}
	return t
}
