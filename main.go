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
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/products"
)

var STOCK_EMOJIS = map[bool]string{
	true:  "00002705",
	false: "0000274C",
}

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
		"%s @ %s, in stock: %s",
		i.name,
		i.price,
		get_emoji(STOCK_EMOJIS[i.is_available()]),
	)
}

func (i item) id() string {
	return fmt.Sprintf("%s: %s", i.product.Name, i.name)
}

func main() {
	telegram_api_ptr := flag.String("api", "", "api token for telegram bot")
	telegram_chat_id_ptr := flag.String("chat", "", "chat id for telegram bot")
	test_ptr := flag.Bool("test", false, "whether to run offline for test purposes")
	update_test_files_ptr := flag.Bool("update-test-files", false, "downloads all test files")
	flag.Parse()

	all_products := products.Products
	if *update_test_files_ptr {
		get_test_files(all_products)
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
	var available_items []item
	var watched_available_items []item
	curr_product := &products.Product{}
	for _, i := range items {
		if i.is_available() {
			available_items = append(available_items, i)
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
				if !already_notified[i.id()] {
					notify_items = append(notify_items, i)
				}
			}
		}
	}

	// Update last_in_stock.txt
	last_in_stock(available_items)

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
		for _, i := range watched_available_items {
			item_names = append(item_names, i.id())
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
		availability: strings.Trim(doc.Find(".bin-signup-dropper button").Text(), " \n"),
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
	f, err := os.Open(p.GetTestFile())
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

func get_test_files(all_products []products.Product) {
	// Make test_pages dir if needed
	if _, err := os.Stat("test_pages"); os.IsNotExist(err) {
		os.Mkdir("test_pages", 0755)
	}

	ch := make(chan bool)
	for _, p := range all_products {
		fmt.Printf("Getting test file for %s...\n", p.Name)
		go get_test_file(p, ch)
	}
	for _, _ = range all_products {
		<-ch
	}
	fmt.Println("Done fetching test files!")
	fmt.Println()
}

func get_test_file(product products.Product, ch chan bool) {
	resp, err := http.Get(product.URL)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	err = ioutil.WriteFile(product.GetTestFile(), body, 0644)
	if err != nil {
		log.Fatal(err)
	}
	ch <- true
}

func last_in_stock(in_stock_now []item) {
	last_in_stock := read_last_in_stock()
	t := time.Now()
	for _, item := range in_stock_now {
		last_in_stock[item.id()] = t.Format("Jan 02, 2006 15:04")
	}

	contents := ""
	for id, timestamp := range last_in_stock {
		contents += fmt.Sprintf("%s :: %s\n", id, timestamp)
	}
	err := ioutil.WriteFile("last_in_stock.txt", []byte(contents), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func read_last_in_stock() map[string]string {
	last_in_stock := map[string]string{}
	file, err := os.Open("last_in_stock.txt")
	if os.IsNotExist(err) {
		last_in_stock = map[string]string{}
	} else if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line_parts := strings.Split(scanner.Text(), " :: ")
		last_in_stock[line_parts[0]] = line_parts[1]
	}
	return last_in_stock
}

func get_emoji(s string) string {
	r, err := strconv.ParseInt(s, 16, 32)
	if err != nil {
		log.Fatal(err)
	}
	return string(r)
}
