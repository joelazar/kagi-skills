SKILLS := kagi-search kagi-fastgpt kagi-summarizer kagi-enrich

.PHONY: build lint test fmt clean

build:
	@for s in $(SKILLS); do \
		echo "=== $$s ===" && cd $$s && go build -o .bin/$$s . && cd ..; \
	done

lint:
	@for s in $(SKILLS); do \
		echo "=== $$s ===" && (cd $$s && golangci-lint run --config ../.golangci.yml); \
	done

test:
	@for s in $(SKILLS); do \
		echo "=== $$s ===" && (cd $$s && go test ./...); \
	done

fmt:
	gofumpt -w -l .

clean:
	@for s in $(SKILLS); do rm -f $$s/.bin/$$s; done
