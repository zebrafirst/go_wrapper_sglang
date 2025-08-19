export GOROOT=/opt/go1.22
export GO111MODULE=on
export PATH="/usr/lib/binutils-2.26/bin:$PATH:/opt/go1.22/bin"
export LD_LIBRARY_PATH=./src/cgo/library
export GOPROXY="https://goproxy.cn,direct"
export GOPRIVATE=git.iflytek.com

printf "machine git.iflytek.com\nlogin {name}\npassword {password}" > ~/.netrc
mv src src.bak
mv go_wrapper_sglang src
cd src
#sed -i 's|D:\\go-local-pkg\\|/home/go-local-pkg/|g' go.mod
go mod tidy
go mod vendor
sleep 15
chmod 755 ./build/fix_import
./build/fix_import ../src
rm ./vendor/golang.org -rf
rm ./vendor/comwrapper -rf
cp -rf ../src.bak/vendor/golang.org ./
cp -rf ../src.bak/comwrapper ./
cd ..

#mv src/vendor/git.iflytek.com/AIaaS/xsf/vendor/git.iflytek.com/AIaaS/finder-go/ src/vendor/git.iflytek.com/AIaaS/xsf/vendor/git.iflytek.com/AIaaS/finder-go.bak
#mv src/vendor/git.iflytek.com/AIaaS/finder-go src/vendor/git.iflytek.com/AIaaS/finder-go.bak
# go build -o mockTest/libwrapper.so -buildmode=plugin src/example/once/wrapper_once.go
# go build -o mockTest/libwrapper.so -buildmode=plugin src/example/stream/wrapper_stream.go

export GO111MODULE=off
export GOPATH=`pwd`

go build -o bin/libwrapper.so -buildmode=plugin src/wrapper/wrapper.go

#mv src/vendor/git.iflytek.com/AIaaS/xsf/vendor/git.iflytek.com/AIaaS/finder-go.bak src/vendor/git.iflytek.com/AIaaS/xsf/vendor/git.iflytek.com/AIaaS/finder-go
#mv src/vendor/git.iflytek.com/AIaaS/finder-go.bak src/vendor/git.iflytek.com/AIaaS/finder-go