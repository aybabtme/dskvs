#!/bin/sh
set -e
mkdir -p doc/cov/
gocov test | gocov-html > doc/cov/dskvs.html
