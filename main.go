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
)

type product struct {
	name         string
	price        string
	availability string
	brand        string
}

func (p product) is_available() bool {
	out_of_stock := map[string]bool{
		"Notify Me":    true,
		"Out of Stock": true,
		"Out of stock": true,
		"OUT OF STOCK": true,
	}
	return !out_of_stock[p.availability]
}

func (p product) String() string {
	return fmt.Sprintf(
		"%s %s @ %s, in stock: %t",
		p.brand,
		p.name,
		p.price,
		p.is_available(),
	)
}

func main() {
	telegram_api_ptr := flag.String("api", "", "api token for telegram bot")
	telegram_chat_id_ptr := flag.String("chat", "", "chat id for telegram bot")
	flag.Parse()

	product_mapping := map[string]string{
		"Rogue Olympic Plates":          "https://www.roguefitness.com/rogue-olympic-plates",
		"Rogue Deep Dish Plates":        "https://www.roguefitness.com/rogue-deep-dish-plates",
		"Rogue Steel Plates":            "https://www.roguefitness.com/rogue-calibrated-lb-steel-plates",
		"Rogue Bumper Plates":           "https://www.roguefitness.com/rogue-hg-2-0-bumper-plates",
		"Rogue Fleck Plate":             "https://www.roguefitness.com/rogue-fleck-plates",
		"Rogue Echo Bumper Plate v2":    "https://www.roguefitness.com/rogue-echo-bumper-plates-with-white-text",
		"Rogue Color Echo Bumper Plate": "https://www.roguefitness.com/rogue-color-echo-bumper-plate",
		"Rep Fitness Cast Iron Plates":  "https://www.repfitness.com/bars-plates/olympic-plates/iron-plates/rep-iron-plates",
		"Rep Fitness Bumper Plates":     "https://www.repfitness.com/bars-plates/olympic-plates/bumper-plates/rep-black-bumper-plates",
	}

	ch := make(chan []product)
	var products []product
	for product_name, url := range product_mapping {
		fmt.Printf("Getting %s...\n", product_name)
		go make_products(ch, url)
	}

	for _, _ = range product_mapping {
		products = append(products, <-ch...)
	}

	fmt.Println("")
	fmt.Println("Available Products:")
	already_notified := get_notified_products()
	var notify_products []product
	var watched_available_products []product
	for _, p := range products {
		if p.is_available() {
			fmt.Println(p)
			if watched(p) {
				watched_available_products = append(watched_available_products, p)
				if !already_notified[p.name] {
					notify_products = append(notify_products, p)
				}
			}
		}
	}

	// Send telegram notification
	if *telegram_api_ptr != "" && *telegram_chat_id_ptr != "" {
		msg := "Watched In Stock Items:\n"
		for _, p := range notify_products {
			msg += fmt.Sprintf("- %s\n", p.name)
		}
		if len(notify_products) > 0 {
			fmt.Println()
			fmt.Println("Sending notification...")
			fmt.Println(msg)
			send_telegram_message(
				*telegram_api_ptr,
				*telegram_chat_id_ptr,
				msg,
			)
		}
		// Store notified products, so as to not re-notify
		var product_names []string
		for _, p := range watched_available_products {
			product_names = append(product_names, p.name)
		}
		err := ioutil.WriteFile("notified_products.txt", []byte(strings.Join(product_names, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func get_notified_products() map[string]bool {
	file, err := os.Open("notified_products.txt")
	if err != nil {
		// The file probably does not exist
		return map[string]bool{}
	}
	defer file.Close()

	notified_products := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		notified_products[strings.Trim(scanner.Text(), "\n")] = true
	}
	return notified_products
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

func watched(p product) bool {
	watched_terms := []string{"45LB", "45 lb"}
	for _, term := range watched_terms {
		if strings.Contains(p.name, term) {
			return true
		}
	}
	return false
}

func make_products(ch chan []product, url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	var products []product
	if strings.Contains(url, "roguefitness") {
		products = make_rogue_products(doc)
	} else if strings.Contains(url, "repfitness") {
		products = make_rep_products(doc)
	}
	ch <- products
}

func make_rep_products(doc *goquery.Document) []product {
	var products []product
	doc.Find(".grouped-items-table tr").Each(func(index int, item *goquery.Selection) {
		p := product{
			name:         strings.Trim(item.Find("td").Eq(0).Text(), " \n"),
			price:        strings.Trim(item.Find(".price").Text(), " \n"),
			availability: strings.Trim(item.Find(".availability").Text(), " \n"),
			brand:        "RepFitness",
		}
		if p.name != "" {
			products = append(products, p)
		}
	})
	return products
}

func make_rogue_products(doc *goquery.Document) []product {
	var products []product
	doc.Find(".grouped-item").Each(func(index int, item *goquery.Selection) {
		p := product{
			name:         item.Find(".item-name").Text(),
			price:        item.Find(".price").Text(),
			availability: strings.Trim(item.Find(".bin-stock-availability").Text(), " \n"),
			brand:        "Rogue",
		}
		products = append(products, p)
	})
	return products
}
