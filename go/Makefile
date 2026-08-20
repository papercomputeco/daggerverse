.PHONY: test
test:
	dagger call --source=./fixtures check-go-mod-tidy
	dagger call --source=./fixtures check-go-vet
	dagger call --source=./fixtures check-go-test
