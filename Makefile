NAMESPACE=vponojko-terraform
NAME=appmixer
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=darwin_arm64

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

clean:
	rm -f ${BINARY}

.PHONY: build install clean 