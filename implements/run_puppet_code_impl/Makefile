BIN_NAME=gcloud_compute_impl

# Make the build dir, and remove any go bins already there
setup:
	mkdir -p ../../output
	rm -rf ../../output/$(BIN_NAME)

# Actually build the thing
build: setup
	cp run_puppet_code_impl.rb ../../output