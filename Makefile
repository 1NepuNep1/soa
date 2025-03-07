test:
	(cd user_service && go test ./... -v)
	(cd api_gateway && go test ./... -v)
