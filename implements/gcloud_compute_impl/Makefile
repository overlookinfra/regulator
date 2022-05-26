GO_PACKAGES=.
GO_MODULE_NAME=github.com/puppetlabs/regulator/implements/gcloud_compute_impl
GO_BIN_NAME=gcloud_compute_impl

# Make the build dir, and remove any go bins already there
setup:
	mkdir -p ../../output
	rm -rf ../../output/$(GO_BIN_NAME)

# Actually build the thing
build: setup
	go mod tidy
	go build -o ../../output/ $(GO_MODULE_NAME)

install:
	go mod tidy
	go install $(GO_MODULE_NAME)

format:
	go fmt $(GO_PACKAGES)