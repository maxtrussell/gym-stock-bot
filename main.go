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
	product      *products.Product
	name         string
	price        string
	availability string
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
		"%s @ %s, in stock: %t",
		i.name,
		i.price,
		i.is_available(),
	)
}

func main() {
	telegram_api_ptr := flag.String("api", "", "api token for telegram bot")
	telegram_chat_id_ptr := flag.String("chat", "", "chat id for telegram bot")
	test_ptr := flag.Bool("test", false, "whether to run offline for test purposes")
	flag.Parse()

	all_products := products.Products
	if *test_ptr {
		all_products = products.TestProducts
	}

	ch := make(chan []item)
	var items []item
	for _, product := range all_products {
		fmt.Printf("Getting %s...\n", product.Name)
		go make_items(ch, product, *test_ptr)
	}

	for _, _ = range all_products {
		items = append(items, <-ch...)
	}

	fmt.Println("")
	fmt.Println("Available Products:")
	already_notified := get_notified_items()
	var notify_items []item
	var watched_available_items []item
	curr_product := &products.Product{}
	for _, i := range items {
		if i.is_available() {
			if i.product.Name != curr_product.Name {
				if curr_product.Name != "" {
					fmt.Println()
				}
				curr_product = i.product
				fmt.Printf("%s:\n", curr_product.Name)
				fmt.Printf("Link: %s\n", curr_product.URL)
			}
			fmt.Printf("- %s\n", i)
			if watched(i) {
				watched_available_items = append(watched_available_items, i)
				if !already_notified[i.name] {
					notify_items = append(notify_items, i)
				}
			}
		}
	}

	// Send telegram notification
	if *telegram_api_ptr != "" && *telegram_chat_id_ptr != "" {
		msg := "Watched In Stock Items:\n"
		curr_product = &products.Product{}
		for _, i := range notify_items {
			if i.product.Name != curr_product.Name {
				if curr_product.Name != "" {
					msg += "\n"
				}
				curr_product = i.product
				msg += fmt.Sprintf("%s:\n", curr_product.Name)
				msg += fmt.Sprintf("Link: %s\n", curr_product.URL)
			}
			msg += fmt.Sprintf("> %s\n", i)
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
		"Ohio Power Bar - Stainless Steel",
		"45LB Rogue Fleck",
		"45LB Rogue Color",
		"1.25LB Rogue Olympic",
		"2.5LB Rogue Olympic",
		"5LB Rogue Olympic",
		"25lb Pair",
	}
	for _, term := range watched_terms {
		if strings.Contains(p.name, term) {
			return true
		}
	}
	return false
}

func make_items(ch chan []item, product products.Product, test bool) {
	var doc *goquery.Document
	var err error
	if test {
		doc = get_test_doc(product)
	} else {
		doc, err = goquery.NewDocument(product.URL)
		if err != nil {
			log.Fatal(err)
		}
	}
	var items []item
	switch product.Brand + ":" + product.Category {
	case "Rogue:multi":
		items = make_rogue_multi_items(doc, product)
	case "Rogue:single":
		items = make_rogue_single_items(doc, product)
	case "RepFitness:rep":
		items = make_rep_items(doc, product)
	}
	ch <- items
}

func make_rep_items(doc *goquery.Document, product products.Product) []item {
	var items []item
	doc.Find("#super-product-table tr").Each(func(index int, selection *goquery.Selection) {
		i := item{
			product:      &product,
			name:         selection.Find(".product-item-name").Text(),
			price:        selection.Find(".price").Text(),
			availability: strings.Trim(selection.Find(".qty-container").Text(), " \n"),
		}
		if i.availability == "" {
			i.availability = strings.Trim(doc.Find(".availability span").Text(), " \n")
		}
		if i.name != "" {
			items = append(items, i)
		}
	})
	return items
}

func make_rogue_single_items(doc *goquery.Document, product products.Product) []item {
	var items []item
	i := item{
		product:      &product,
		name:         doc.Find(".product-title").Text(),
		price:        doc.Find(".price").Text(),
		availability: strings.Trim(doc.Find(".product-options-bottom button").Text(), " \n"),
	}
	items = append(items, i)
	return items
}

func make_rogue_multi_items(doc *goquery.Document, product products.Product) []item {
	var items []item
	doc.Find(".grouped-item").Each(func(index int, selection *goquery.Selection) {
		i := item{
			product:      &product,
			name:         selection.Find(".item-name").Text(),
			price:        selection.Find(".price").Text(),
			availability: strings.Trim(selection.Find(".bin-stock-availability").Text(), " \n"),
		}
		items = append(items, i)
	})
	return items
}

func get_test_doc(p products.Product) *goquery.Document {
	f, err := os.Open(p.TestFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
