#!/usr/bin/env bash

set -e

BOOK_DIR=$(readlink -f $1)

task build-single
PATH=${PWD}/dist:$PATH mdbook build ${BOOK_DIR}
docker run --rm -u $(id -u):$(id -g) -v ${BOOK_DIR}/book/asciidoc/:/documents/ asciidoctor/docker-asciidoctor asciidoctor-pdf -a allow-uri-read output.adoc
