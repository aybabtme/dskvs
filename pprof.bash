#! /bin/sh
set -e
mkdir -p doc/pprof/
go test -cpuprofile doc/pprof/cpu.pprof
go test -memprofile doc/pprof/mem.pprof
go tool pprof --svg dskvs.test doc/pprof/mem.pprof > doc/pprof/mem_profile.svg
go tool pprof --svg dskvs.test doc/pprof/cpu.pprof > doc/pprof/cpu_profile.svg
rm dskvs.test
