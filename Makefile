VERSION = $(shell cat .version)
GHT = $(GITHUB_TOKEN)

release: postfix_exporter
	ghr  --username baldguysoftware --token ${GITHUB_TOKEN} --replace ${VERSION} postfix_exporter


pf2s3: 
	go get -t ./...
	go build 

test:
	go test
	go vet

