.PHONY: load-util

OS := $(shell uname)

ifeq ($(OS), Darwin)
    load-util:
		wget https://obs-community-intl.obs.ap-southeast-1.myhuaweicloud.com/obsutil/current/obsutil_darwin_amd64.tar.gz
		tar -xzvf obsutil_darwin_amd64.tar.gz
		cd obsutil_darwin_amd64_5.5.12 && sudo mv obsutil /usr/local/bin
		obsutil version
		sudo rm -rf obsutil_darwin_amd64_5.5.12 obsutil_darwin_amd64.tar.gz
else ifeq ($(OS), Linux)
    load-util:
		wget https://obs-community-intl.obs.ap-southeast-1.myhuaweicloud.com/obsutil/current/obsutil_linux_amd64.tar.gz
		tar -xzvf obsutil_linux_amd64.tar.gz
		cd obsutil_linux_amd64_5.5.12 && sudo mv obsutil /usr/local/bin
		obsutil version
		sudo rm -rf obsutil_linux_amd64_5.5.12 obsutil_linux_amd64.tar.gz
else
    load-util:
        @echo "Unsupported operating system: $(OS)"
endif

.PHONY: start
start:
		go run cmd/main.go