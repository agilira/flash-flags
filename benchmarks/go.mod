module github.com/agilira/flash-flags/benchmarks

go 1.23.11

require (
	github.com/agilira/flash-flags v0.0.0
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/jessevdk/go-flags v1.6.1
	github.com/spf13/pflag v1.0.10
)

require (
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)

replace github.com/agilira/flash-flags => ../
