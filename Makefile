.PHONY: statik
statik:
	statik -src ./templates
	gofmt -w ./statik/statik.go
