ARG base_image

FROM $base_image as builder
WORKDIR /go/src/github.com/packagrio/releasr
COPY . .
RUN go mod vendor && go build -o /go/bin/packagr-releasr cmd/releasr/releasr.go

##################################################
##
## Golang
##
##################################################
FROM $base_image as runtime
MAINTAINER Jason Kulatunga <jason@thesparktree.com>
ENV PACKAGR_ENGINE_GIT_AUTHOR_NAME="packagrio-bot" \
    PACKAGR_ENGINE_GIT_AUTHOR_EMAIL="packagr-io[bot]@users.noreply.github.com" \
    PACKAGR_PACKAGE_TYPE=golang

WORKDIR /srv/packagr

RUN apt-get update && apt-get install -y --no-install-recommends \
 	apt-transport-https \
    ca-certificates \
    git \
    curl \
	locales \
	&& rm -rf /var/lib/apt/lists/* \
	&& locale-gen en_US.UTF-8

ENV PATH="/srv/packagr:${PATH}" \
	SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
	LANG=en_US.UTF-8 \
	LANGUAGE=en_US.UTF-8 \
	LC_ALL=en_US.UTF-8

COPY --from=builder /go/bin/packagr-releasr .


CMD "packagr-releasr"
