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
			timeSince := parseTime(rows[i-1].Timestamp).Sub(parseTime(r.Timestamp))
			if r.InStock {
				time_in_stock += timeSince.Seconds()
			} else {
				time_out_of_stock += timeSince.Seconds()
			}
		}
	}
	// update to present time
	timeSince := time.Now().Sub(parseTime(rows[0].Timestamp))
	if rows[0].InStock {
		time_in_stock += timeSince.Seconds()
	} else {
		time_out_of_stock += timeSince.Seconds()
	}

	fmt.Printf("Printing report for \"%s\"\n", id)
	fmt.Printf("Data since: %s\n", since)
	fmt.Printf("In stock: %t\n", rows[0].InStock)

	// 1. Last in stock/Last out of stock
	fmt.Printf("Last in stock: %s\n", last_in_stock)
	fmt.Printf("Last out of stock: %s\n", last_out_of_stock)

	// 2. Number of times in stock
	fmt.Printf("Times in stock: %d\n", times_in_stock)
	fmt.Printf("Times out of stock: %d\n", times_out_of_stock)

	// 3. Average days in/out of stock
	avg_in_stock := time_in_stock / float64(times_in_stock)
	avg_out_of_stock := time_out_of_stock / float64(times_out_of_stock)
	fmt.Printf("Time in stock: %s\n", formatSeconds(int(time_in_stock)))
	fmt.Printf("Time out of stock: %s\n", formatSeconds(int(time_out_of_stock)))
	fmt.Printf("Avg time in stock: %s\n", formatSeconds(int(time_in_stock)/times_in_stock))
	fmt.Printf("Avg time out of stock: %s\n", formatSeconds(int(time_out_of_stock)/times_out_of_stock))

	// 4. Predicted next in/out of stock
	var predicted float64
	if rows[0].InStock {
		predicted = avg_in_stock - timeSince.Seconds()
	} else {
		predicted = avg_out_of_stock - timeSince.Seconds()
	}
	fmt.Printf("Predicted next stock change: %s\n", formatSeconds(int(predicted)))
}

func formatSeconds(s int) string {
	// Converts seconds to a xhyy format
	secs_per_day := 24 * 60 * 60
	days := s / secs_per_day
	hours := (s % secs_per_day) / 3600
	mins := (s % 3600) / 60
	msg := fmt.Sprintf("%dh%d", hours, mins)
	if days > 0 {
		msg = fmt.Sprintf("%d days %s", days, msg)
	}
	return msg
}

func parseTime(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		log.Fatal(err)
	}
	return t
}
