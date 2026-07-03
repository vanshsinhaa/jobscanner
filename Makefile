.PHONY: run watch report baseline build vet clean

run:            ## full pipeline: scrape -> classify -> README + jobs.json
	go run main.go

watch:          ## local daemon mode, 15m interval
	go run main.go --watch --interval=15m

report:         ## 7-day coverage per target company
	go run main.go --target-report

baseline:       ## test every scraper in isolation (no DB/README writes)
	go run ./cmd/baseline

build:
	go build ./...

vet:
	go vet ./...

clean:          ## remove local scrape state (safe: CI state lives in the data repo)
	rm -f local_data/jobs.db local_data/jobs.db-journal
