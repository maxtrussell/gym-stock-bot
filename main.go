package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/products"
)

type item struct {
	name         string
	price        string
	availability string
	brand        string
}

func (i item) is_available() bool {
	out_of_stock := map[string]bool{
		"Notify Me":    true,
		"Out of Stock": true,
		"Out of stock": true,
		"OUT OF STOCK": true,
	}
	return !out_of_stock[i.availability]
}

func (i item) String() string {
	return fmt.Sprintf(
		"%s %s @ %s, in stock: %t",
		i.brand,
		i.name,
		i.price,
		i.is_available(),
	)
}

func main() {
	telegram_api_ptr := flag.String("api", "", "api token for telegram bot")
	telegram_chat_id_ptr := flag.String("chat", "", "chat id for telegram bot")
	flag.Parse()

	ch := make(chan []item)
	var items []item
	for _, product := range products.Products {
		fmt.Printf("Getting %s...\n", product.Name)
		go make_items(ch, product)
	}

	for _, _ = range products.Products {
		items = append(items, <-ch...)
	}

	fmt.Println("")
	fmt.Println("Available Products:")
	already_notified := get_notified_items()
	var notify_items []item
	var watched_available_items []item
	for _, p := range items {
		if p.is_available() {
			fmt.Println(p)
			if watched(p) {
				watched_available_items = append(watched_available_items, p)
				if !already_notified[p.name] {
					notify_items = append(notify_items, p)
				}
			}
		}
	}

	// Send telegram notification
	if *telegram_api_ptr != "" && *telegram_chat_id_ptr != "" {
		msg := "Watched In Stock Items:\n"
		for _, p := range notify_items {
			msg += fmt.Sprintf("> %s\n", p)
		}
		if len(notify_items) > 0 {
			fmt.Println()
			fmt.Println("Sending notification...")
			fmt.Println(msg)
			send_telegram_message(
				*telegram_api_ptr,
				*telegram_chat_id_ptr,
				msg,
			)
		}
		// Store notified items, so as to not re-notify
		var item_names []string
		for _, p := range watched_available_items {
			item_names = append(item_names, p.name)
		}
		err := ioutil.WriteFile("notified_items.txt", []byte(strings.Join(item_names, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func get_notified_items() map[string]bool {
	file, err := os.Open("notified_items.txt")
	if err != nil {
		// The file probably does not exist
		return map[string]bool{}
	}
	defer file.Close()

	notified_items := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		notified_items[strings.Trim(scanner.Text(), "\n")] = true
	}
	return notified_items
}

func send_telegram_message(api_token, chat_id, msg string) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", api_token)
	data := url.Values{
		"chat_id": {chat_id},
		"text":    {msg},
	}
	_, err := http.PostForm(endpoint, data)
	if err != nil {
		log.Fatal(err)
	}
}

func watched(p item) bool {
	watched_terms := []string{
		"45LB", "45 lb",
		"Ohio Power Bar - Stainless Steel",
	}
	for _, term := range watched_terms {
		if strings.Contains(p.name, term) {
			return true
		}
	}
	return false
}

func make_items(ch chan []item, product products.Product) {
	doc, err := goquery.NewDocument(product.URL)
	if err != nil {
		log.Fatal(err)
	}
	var items []item
	switch product.Brand + ":" + product.Category {
	case "Rogue:multi":
		items = make_rogue_multi_items(doc)
	case "Rogue:single":
		items = make_rogue_single_items(doc)
	case "RepFitness:rep":
		items = make_rep_items(doc)
	}
	ch <- items
}

func make_rep_items(doc *goquery.Document) []item {
	var items []item
	doc.Find(".grouped-items-table tr").Each(func(index int, selection *goquery.Selection) {
		i := item{
			name:         strings.Trim(selection.Find("td").Eq(0).Text(), " \n"),
			price:        strings.Trim(selection.Find(".price").Text(), " \n"),
			availability: strings.Trim(selection.Find(".availability").Text(), " \n"),
			brand:        "RepFitness",
		}
		if i.name != "" {
			items = append(items, i)
		}
	})
	return items
}

func make_rogue_single_items(doc *goquery.Document) []item {
	var items []item
	i := item{
		name:         doc.Find(".product-title").Text(),
		price:        doc.Find(".price").Text(),
		availability: strings.Trim(doc.Find(".product-options-bottom button").Text(), " \n"),
		brand:        "Rogue",
	}
	items = append(items, i)
	return items
}

func make_rogue_multi_items(doc *goquery.Document) []item {
	var items []item
	doc.Find(".grouped-item").Each(func(index int, selection *goquery.Selection) {
		i := item{
			name:         selection.Find(".item-name").Text(),
			price:        selection.Find(".price").Text(),
			availability: strings.Trim(selection.Find(".bin-stock-availability").Text(), " \n"),
			brand:        "Rogue",
		}
		items = append(items, i)
	})
	return items
}
