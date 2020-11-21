package rogue

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/models/item"
	"github.com/maxtrussell/gym-stock-bot/models/product"
)

func MakeRogue(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	switch product.Category {
	case "multi":
		items = makeRogueMulti(doc, product)
	case "single":
		items = makeRogueSingle(doc, product)
	}
	return items
}

func makeRogueSingle(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	i := item.Item{
		Product:      &product,
		Name:         doc.Find(".product-title").Text(),
		Price:        doc.Find(".price").Text(),
		Availability: strings.Trim(doc.Find(".bin-signup-dropper button").Text(), " \n"),
	}
	items = append(items, i)
	return items
}

func makeRogueMulti(doc *goquery.Document, product product.Product) []item.Item {
	var items []item.Item
	doc.Find(".grouped-item").Each(func(index int, selection *goquery.Selection) {
		i := item.Item{
			Product:      &product,
			Name:         selection.Find(".item-name").Text(),
			Price:        selection.Find(".price").Text(),
			Availability: strings.Trim(selection.Find(".bin-stock-availability").Text(), " \n"),
		}
		items = append(items, i)
	})
	return items
}
