.PHONY: load-util

OS := $(shell uname)

ifeq ($(OS), Darwin)
    load-util:
		wget https://obs-community-intl.obs.ap-southeast-1.myhuaweicloud.com/obsutil/current/obsutil_darwin_amd64.tar.gz
		tar -xzvf obsutil_darwin_amd64.tar.gz
		cd obsutil_darwin_amd64_5.5.12 && sudo mv obsutil /usr/local/bin
		obsutil version
		sudo rm -rf obsutil_darwin_amd64_5.5.12 obsutil_darwin_amd64.tar.gz
		wget https://releases.hashicorp.com/terraform/1.0.10/terraform_1.0.10_linux_amd64.zip
		unzip terraform_1.0.10_linux_amd64.zip
		sudo mv terraform /usr/local/bin/
		sudo rm -rf terraform_1.0.10_linux_amd64.zip
else ifeq ($(OS), Linux)
    load-util:
		wget https://obs-community-intl.obs.ap-southeast-1.myhuaweicloud.com/obsutil/current/obsutil_linux_amd64.tar.gz
		tar -xzvf obsutil_linux_amd64.tar.gz
		cd obsutil_linux_amd64_5.5.12 && sudo mv obsutil /usr/local/bin
		obsutil version
		sudo rm -rf obsutil_linux_amd64_5.5.12 obsutil_linux_amd64.tar.gz
		wget https://releases.hashicorp.com/terraform/1.0.10/terraform_1.0.10_linux_amd64.zip
		unzip terraform_1.0.10_linux_amd64.zip
		sudo mv terraform /usr/local/bin/
		sudo rm -rf terraform_1.0.10_linux_amd64.zip

else
    load-util:
        @echo "Unsupported operating system: $(OS)"
endif

.PHONY: start
start:
	rm -rf *.log
	export https_proxy=172.23.144.4:3128 && go run cmd/main.go