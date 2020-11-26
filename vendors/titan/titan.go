package titan

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/models/item"
	"github.com/maxtrussell/gym-stock-bot/models/product"
)

func MakeTitan(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	switch product.Category {
	case "rack":
		items = makeTitanRack(doc, product)
	}
	return items
}
