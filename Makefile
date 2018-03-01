apply-terraform: api.zip
	./apply

api.zip: main schema.graphql rds-ca.pem
	rm -f api.zip
	zip api.zip main schema.graphql rds-ca.pem

main: api/main.go api/resolver.go api/main_test.go schema.graphql
	cd api && golint
	cd api && environment=local go test
	cd api && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../main

rds-ca.pem:
	curl -L 'https://s3.amazonaws.com/rds-downloads/rds-ca-2015-root.pem' -o "$@"
