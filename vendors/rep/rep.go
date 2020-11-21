package rep

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/models/item"
	"github.com/maxtrussell/gym-stock-bot/models/product"
)

func MakeRep(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	switch product.Category {
	case "multi":
		items = makeRepMulti(doc, product)
	}
	return items
}

func makeRepMulti(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	doc.Find("#super-product-table tr").Each(func(index int, selection *goquery.Selection) {
		i := item.Item{
			Product:      &product,
			Name:         selection.Find(".product-item-name").Text(),
			Price:        selection.Find(".price").Text(),
			Availability: strings.Trim(selection.Find(".qty-container").Text(), " \n"),
		}
		if i.Availability == "" {
			i.Availability = strings.Trim(doc.Find(".availability span").Text(), " \n")
		}
		if i.Name != "" {
			items = append(items, i)
		}
	})
	return items
}
