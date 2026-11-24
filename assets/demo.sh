#!/usr/bin/env bash
# demo.sh: the session recorded in assets/demo.gif
#
# Regenerate (asciinema 2.x and agg):
#   asciinema rec --cols 104 --rows 25 -c assets/demo.sh demo.cast
#   agg --theme monokai --font-size 16 --idle-time-limit 3 demo.cast assets/demo.gif
#
# 104x25 renders close to 16:9.

set -e
repo="$(cd "$(dirname "$0")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

(cd "$repo/examples/server" && go build -o "$work/server" .)
cd "$work"
export PATH="$work:$PATH"
printf '{\n  "port": 9000,\n  "timeout": "45s"\n}\n' > server.json

say() { printf '\033[38;5;244m# %s\033[0m\n' "$1"; sleep 0.8; }

run() {
	printf '\033[1;32m$\033[0m '
	for ((i = 0; i < ${#1}; i++)); do
		printf '%s' "${1:i:1}"
		sleep 0.035
	done
	sleep 0.4
	echo
	eval "$1" || true
	echo
	sleep "${2:-1.8}"
}

clear
say "flash-flags: command line, environment and config file in one parser"
run "server --help" 3.5

clear
say "a config file sets the base values..."
run "cat server.json" 1.2
run "server"
say "...an environment variable outranks it..."
run "SERVER_PORT=7000 server" 2.5

clear
say "...and the command line outranks both: short, combined, repeated"
run "SERVER_PORT=7000 server -vp 3000 --tag api --tag edge" 2.5
say "validators run whichever layer supplied the value"
run "server --port 99999" 3
