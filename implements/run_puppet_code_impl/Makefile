BIN_NAME=run_puppet_code_impl

# Make the build dir, and remove any go bins already there
setup:
	mkdir -p ../../output
	rm -rf ../../output/$(BIN_NAME).rb
	rm -rf ../../output/$(BIN_NAME).yaml

# Actually build the thing
build: setup
	cp $(BIN_NAME).rb ../../output
	cp $(BIN_NAME).yaml ../../output