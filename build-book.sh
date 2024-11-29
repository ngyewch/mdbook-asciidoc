#!/usr/bin/env bash

set -e

BOOK_DIR=$(readlink -f $1)

PATH=${PWD}/dist:$PATH mdbook build ${BOOK_DIR}

mkdir -p ${BOOK_DIR}/book/pdf
docker run --rm \
    -u $(id -u):$(id -g) \
    -v ${BOOK_DIR}/book/asciidoc/:/documents/ \
    -v ${BOOK_DIR}/book/pdf/:/output/ \
    asciidoctor/docker-asciidoctor \
    asciidoctor-pdf -a allow-uri-read -D /output/ output.adoc
