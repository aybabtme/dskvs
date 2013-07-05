#!/bin/sh
gocov test | gocov-html > doc/cov/dskvs.html
