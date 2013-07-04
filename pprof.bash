#! /bin/sh
go test -cpuprofile doc/cpu.pprof
go test -memprofile doc/mem.pprof
go tool pprof --svg dskvs.test doc/mem.pprof > doc/mem_profile.svg
go tool pprof --svg dskvs.test doc/cpu.pprof > doc/cpu_profile.svg
rm dskvs.test
