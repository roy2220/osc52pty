#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail # -o xtrace

OSC52_BEGIN="\x1b]52;c;"
OSC52_END="\x07"

if [[ -v TERM && ${TERM} == screen* ]]; then
    if [[ -v TMUX ]]; then
        OSC52_BEGIN="\x1bPtmux;\x1b${OSC52_BEGIN}"
        OSC52_END="${OSC52_END}\x1b\\"
    else
        OSC52_BEGIN="\x1bP;${OSC52_BEGIN}"
        OSC52_END="${OSC52_END}\x1b\\"
    fi
fi

sed_escape() {
    echo -e "${1}" | sed 's/[]\/$*.^[]/\\&/g'
}

TTY=/dev/tty

if ! echo -n > "${TTY}" 2>&1; then
    if [[ -v TERM && ${TERM} == screen* ]]; then
        if [[ -v TMUX ]]; then
            TTY=$(tmux display -p '#{pane_tty}')
        fi
    fi
fi

openssl base64 -A | sed -e 's/^/'"$(sed_escape "${OSC52_BEGIN}")"'/' -e 's/$/'"$(sed_escape "${OSC52_END}")"'/' > "${TTY}"
