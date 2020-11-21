package item

import (
	"fmt"
	"log"
	"strconv"

	"github.com/maxtrussell/gym-stock-bot/models/product"
)

var STOCK_EMOJIS = map[bool]string{
	true:  "00002705",
	false: "0000274C",
}

type Item struct {
	Product      *product.Product
	Name         string
	Price        string
	Availability string
}

func (i Item) IsAvailable() bool {
	out_of_stock := map[string]bool{
		"Notify Me":    true,
		"Out of Stock": true,
		"Out of stock": true,
		"OUT OF STOCK": true,
	}
	return !out_of_stock[i.Availability]
}

func (i Item) String() string {
	return fmt.Sprintf(
		"%s @ %s, in stock: %s",
		i.Name,
		i.Price,
		get_emoji(STOCK_EMOJIS[i.IsAvailable()]),
	)
}

func (i Item) ID() string {
	return fmt.Sprintf("%s: %s", i.Product.Name, i.Name)
}

func get_emoji(s string) string {
	r, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		log.Fatal(err)
	}
	return string(r)
}
