
build:
	go build .

clean:
	rm notified_items.txt
	mv db.sqlite db.sqlite.bak
