# Sir I have no idea
build:
	go build -a -ldflags '-w -extldflags "-static"' -o bin/qiniu-cert-sync

dockerbuild:
	docker build -t bohrasd/qiniu-cert-sync .

dockerpush:
	docker push bohrasd/qiniu-cert-sync

fmt:
	go mod tidy
	gofmt -s -w .
