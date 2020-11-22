package rogue

import (
	"encoding/json"
	"log"
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
	case "script":
		items = makeFromScript(doc, product, "RogueColorSwatches")
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
			Name:         strings.TrimSpace(selection.Find(".item-name").Text()),
			Price:        selection.Find(".price").Text(),
			Availability: strings.Trim(selection.Find(".bin-stock-availability").Text(), " \n"),
		}
		items = append(items, i)
	})
	return items
}

func makeFromScript(doc *goquery.Document, product product.Product, script_name string) []item.Item {
	var items []item.Item
	// Find json blob to parse
	selection := doc.Find("script[type='text/javascript']")
	for _, node := range selection.Nodes {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			i := strings.Index(c.Data, script_name)
			if i == -1 {
				continue
			}
			for _, line := range strings.Split(c.Data, "\n") {
				stripped_line := strings.Trim(line, " ")
				if strings.HasPrefix(stripped_line, "{") {
					items = append(
						items,
						parseColorSwatchJson(stripped_line[:len(stripped_line)-1], product)...,
					)
				}
			}
		}
	}
	return items
}

// helper function for makeFromScript
func parseColorSwatchJson(color_swatch string, product product.Product) []item.Item {
	var top_level map[string]interface{}
	err := json.Unmarshal([]byte(color_swatch), &top_level)
	if err != nil {
		log.Fatal(err)
	}

	// 1. Get options
	var item_options []map[string]interface{}
	attributes := top_level["attributes"].(map[string]interface{})
	for _, val := range attributes {
		options, _ := val.(map[string]interface{})["options"]
		for _, option := range options.([]interface{}) {
			if additional_option, ok := option.(map[string]interface{})["additional_options"]; ok {
				for _, item_info := range additional_option.(map[string]interface{}) {
					item_options = append(item_options, item_info.(map[string]interface{}))
				}
			}
		}
	}

	// 2. Convert options to items
	var items []item.Item
	for _, option := range item_options {
		var availability string
		if option["isInStock"].(bool) {
			availability = "In stock"
		} else {
			availability = "Out of stock"
		}
		i := item.Item{
			Product:      &product,
			Name:         option["realLabel"].(string)[3:],
			Price:        "$" + strings.TrimSuffix(option["bin_price"].(string), "00"),
			Availability: availability,
		}
		items = append(items, i)
	}

	return items
}
