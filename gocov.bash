#!/bin/sh
gocov test -short . | gocov-html > doc/cov/dskvs.html
