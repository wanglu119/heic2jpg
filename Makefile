mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
cur_makefile_path := $(patsubst %/Makefile, %, $(mkfile_path))
#window "/" transe to "\"
cur_path := $(patsubst %/, %\,$(cur_makefile_path))
export GOPATH:= ${cur_path}
export GO111MODULE=auto

default: heic2jpg

heic2jpg: 
	go install ./src/github.com/wanglu119/heic2jpg
	go build -o ./bin/heic2jpg ./src/github.com/wanglu119/heic2jpg

heic2jpg-win: 
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix \
	go build -o ./bin/heic2jpg_x86_64.exe -ldflags '-extldflags "-static"' \
	./src/github.com/wanglu119/heic2jpg

release:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build -o ./bin/heic2jpg_linux_amd64 \
	./src/github.com/wanglu119/heic2jpg

	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix \
	go build -o ./bin/heic2jpg_windows_x86_64.exe -ldflags '-extldflags "-static"' \
	./src/github.com/wanglu119/heic2jpg