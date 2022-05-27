GO_PACKAGES=. ./connection ./local ./localexec ./localfile ./operation ./operparse ./remote ./render ./rgerror ./sanitize ./validator ./version
GO_MODULE_NAME=github.com/puppetlabs/regulator
GO_BIN_NAME=regulator

# Make the build dir, and remove any go bins already there
setup:
	mkdir -p output/
	rm -rf output/$(GO_BIN_NAME)

# Actually build the thing
build-regulator: setup
	go mod tidy
	go build -o output/ $(GO_MODULE_NAME)

build-implements:
	cd implements && \
	for DIR in $$(ls); do \
		cd $$DIR && \
		make build && \
		cd ..; \
	done && \
	cd .. && \
	git checkout -- implements/**/go.mod

build: build-regulator build-implements

install:
	go mod tidy
	go install $(GO_MODULE_NAME)

# Build it before publishing to make sure this publication won't be broken
publish: build
ifndef NEW_VERSION
	echo "Cannot publish, no tag provided. Set NEW_VERSION to new tag"
else
	git tag -a $(NEW_VERSION) -m "Version $(NEW_VERSION)";
	git push
	git push --tags
endif

format:
	go fmt $(REGULATOR_GO_PACKAGES)