PLATFORMS ?= linux/amd64 linux/arm64

.PHONY: test
test:
	@for platform in $(PLATFORMS); do \
		dagger call \
			--source=./fixtures \
			with-build-arg --name=MESSAGE --value=expected \
			build --platform=$$platform \
			sync >/dev/null || exit 1; \
	done
