package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/maxtrussell/gym-stock-bot/analytics"
	"github.com/maxtrussell/gym-stock-bot/database"
	"github.com/maxtrussell/gym-stock-bot/models/item"
	"github.com/maxtrussell/gym-stock-bot/models/product"
	"github.com/maxtrussell/gym-stock-bot/telegram"
	"github.com/maxtrussell/gym-stock-bot/vendors/rep"
	"github.com/maxtrussell/gym-stock-bot/vendors/rogue"
	"github.com/maxtrussell/gym-stock-bot/web"
)

func main() {
	telegram_api_ptr := flag.String("api", "", "api token for telegram bot")
	telegram_chat_id_ptr := flag.String("chat", "", "chat id for telegram bot")
	telegram_server := flag.Bool("server", false, "whether or not to be a telegram server")
	test_ptr := flag.Bool("test", false, "whether to run offline for test purposes")
	update_test_files_ptr := flag.Bool("update-test-files", false, "downloads all test files")
	update_db_ptr := flag.Bool("update-db", false, "whether to update the stock db")
	analytics_ptr := flag.String("analyze", "", "item id name to analyze")
	flag.Parse()

	start_time := time.Now()
	fmt.Printf("Current time: %s\n", start_time)

	if *telegram_server {
		go telegram.ListenAndServe(*telegram_api_ptr)
		web.ListenAndServe()
		return
	}

	if *analytics_ptr != "" {
		db := database.Setup()
		analytics.ItemReport(db, *analytics_ptr)
		return
	}

	all_products := product.Products
	if *update_test_files_ptr {
		get_test_files(all_products)
	}

	ch := make(chan []item.Item)
	var items []item.Item
	for _, product := range all_products {
		fmt.Printf("Getting %s...\n", product.Name)
		go make_items(ch, product, *test_ptr)
	}

	for _, _ = range all_products {
		items = append(items, <-ch...)
	}

	watched_terms := read_watched()

	fmt.Println("")
	fmt.Println("Available Products:")
	already_notified := get_notified_items()
	var notify_items []item.Item
	var available_items []item.Item
	var watched_available_items []item.Item
	curr_product := &product.Product{}
	for _, i := range items {
		if i.IsAvailable() {
			available_items = append(available_items, i)
			if i.Product.Name != curr_product.Name {
				if curr_product.Name != "" {
					fmt.Println()
				}
				curr_product = i.Product
				fmt.Printf("%s:\n", curr_product.Name)
				fmt.Printf("Link: %s\n", curr_product.URL)
			}
			fmt.Printf("- %s\n", i)
			if watched(i, watched_terms) {
				watched_available_items = append(watched_available_items, i)
				if !already_notified[i.ID()] {
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
		curr_product = &product.Product{}
		for _, i := range notify_items {
			if i.Product.Name != curr_product.Name {
				if curr_product.Name != "" {
					msg += "\n"
				}
				curr_product = i.Product
				msg += fmt.Sprintf("%s:\n", curr_product.Name)
				msg += fmt.Sprintf("Link: %s\n", curr_product.URL)
			}
			msg += fmt.Sprintf("> %s\n", i)
		}
		if len(notify_items) > 0 {
			fmt.Println()
			fmt.Println("Sending notification...")
			fmt.Println(msg)
			telegram.SendMessage(
				*telegram_api_ptr,
				*telegram_chat_id_ptr,
				msg,
			)
		}
		// Store notified items, so as to not re-notify
		var item_names []string
		for _, i := range watched_available_items {
			item_names = append(item_names, i.ID())
		}
		err := ioutil.WriteFile("notified_items.txt", []byte(strings.Join(item_names, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Update the stock db
	if *update_db_ptr {
		db := database.Setup()
		database.UpdateStock(db, items)
	}

	end_time := time.Now()
	fmt.Println()
	fmt.Printf("Completed in %.2f seconds\n", end_time.Sub(start_time).Seconds())
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

func watched(i item.Item, watched_terms []string) bool {
	for _, term := range watched_terms {
		re := regexp.MustCompile(term)
		if re.FindStringIndex(i.ID()) != nil {
			return true
		}
	}
	return false
}

func make_items(ch chan []item.Item, product product.Product, test bool) {
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
	var items []item.Item
	switch product.Brand {
	case "Rogue":
		items = rogue.MakeRogue(doc, product)
	case "RepFitness":
		items = rep.MakeRep(doc, product)
	}
	ch <- items
}

func get_test_doc(p product.Product) *goquery.Document {
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

func get_test_files(all_products []product.Product) {
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

func get_test_file(product product.Product, ch chan bool) {
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

func last_in_stock(in_stock_now []item.Item) {
	last_in_stock := read_last_in_stock()
	t := time.Now()
	for _, item := range in_stock_now {
		last_in_stock[item.ID()] = t.Format("Jan 02, 2006 15:04")
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

func read_watched() []string {
	watched := []string{}
	file, err := os.Open("watched.txt")
	if os.IsNotExist(err) {
		return watched
	} else if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		watched = append(watched, scanner.Text())
	}
	return watched
}
