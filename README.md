# osc52pty

[![Build Status](https://travis-ci.com/roy2220/osc52pty.svg?branch=master)](https://travis-ci.com/roy2220/osc52pty) [![Coverage Status](https://codecov.io/gh/roy2220/osc52pty/branch/master/graph/badge.svg)](https://codecov.io/gh/roy2220/osc52pty)

osc52pty recognizes OSC 52 sequences and pipes the content to `pbcopy`.

If you're a Mac user who loves Terminal.app, you must be envious of iTerm2 users who
can easily send a text to local clipboard from a remote side.

How the magic works? OSC 52 is one of [Xterm Control Sequences](https://www.xfree86.org/current/ctlseqs.html),
which is designated for clipboard setting. Once a terminal supporting OSC 52 catches a
text in the form of OSC 52 from the output, instead of printing the text onto the screen,
it decodes the text first and then sends the content to the system clipboard.

Although Terminal.app do NOT support OSC 52, here is the workaround for it.

## Installation

Go 1.18 or newer is required. Install the latest release with:

```bash
go install github.com/roy2220/osc52pty@latest
```

The binary is installed in `GOBIN`, or in `GOPATH/bin` when `GOBIN` is not
set:

```bash
ls -lh "$(go env GOPATH)/bin/osc52pty"
```

## Usage

Launch your default shell with `osc52pty` to add OSC 52 clipboard support:

```bash
osc52pty
```

You can also provide a shell or another command explicitly, for example
`osc52pty /bin/zsh`.

Within the shell launched, send a OSC 52 sequence to testify:

```bash
printf "\e]52;c;%s\a" "$(echo -n 'THE TEXT TO COPY' | openssl base64 -A)"
```

Now the system clipboard is set to `THE TEXT TO COPY`.

### Remote tmux over SSH

Start `osc52pty` **on the Mac**, then connect to the remote host from the
wrapped shell:

```bash
osc52pty
ssh user@example.com
```

On the remote host, tmux must be told that its outer terminal can set the
clipboard. For tmux 3.2 or newer, add the following to `~/.tmux.conf`:

```tmux
set -s set-clipboard external
set -as terminal-features ',xterm-256color:clipboard'
```

Reload the configuration, then verify that tmux has the `Ms` clipboard
capability:

```bash
tmux source-file ~/.tmux.conf
tmux info | grep 'Ms:'
```

`Ms` should contain an OSC 52 sequence instead of showing `[missing]`. Text
copied in tmux copy mode will then be written to the Mac system clipboard by
`osc52pty`, including across SSH. For example, with the default prefix and vi
copy-mode keys: press `Ctrl-b [`, start a selection with `Space`, move to the
end, then press `Enter`. The copied text can be pasted into a Mac application
with `Command-V`.

For a direct protocol test through tmux, use:

```bash
printf "\ePtmux;\e\e]52;c;%s\a\e\\" "$(echo -n 'THE TEXT TO COPY' | openssl base64 -A)"
```

## See also

- [remote-pbcopy-iterm2](https://github.com/skaji/remote-pbcopy-iterm2)
- [tmux clipboard documentation](https://github.com/tmux/tmux/wiki/Clipboard)
