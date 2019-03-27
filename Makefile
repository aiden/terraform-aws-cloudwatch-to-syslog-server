install:
	# This takes a while:
	(cd test; dep ensure)

build:
	# This takes a while:
	(cd lambda; yarn --frozen-lockfile)
	rm -f lambda.zip
	(cd lambda; zip --quiet ../lambda.zip -r *)

test: test/**/* examples/**/* *.tf
	(cd test; go test)
