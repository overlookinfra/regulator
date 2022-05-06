REGULATOR_GO_PACKAGES=. ./connection ./language ./local ./remote ./rgerror ./utils ./validator ./version

# Make the build dir, and remove anything already inside it
setup:
	mkdir -p output
	rm -rf output/*

# Actually build the thing
build: setup
	go mod tidy
	go build -o output/ github.com/mcdonaldseanp/regulator

install:
	go mod tidy
	go install github.com/mcdonaldseanp/regulator

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