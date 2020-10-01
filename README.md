# osc52pty

osc52pty recognizes OSC 52 sequences and pipes the content to `pbcopy`.

If you're a Mac user who loves Terminal.app, you must be envious of iTerm2 users who
can easily send a text to local clipboard from a remote side.

How the magic works? OSC 52 is one of [Xterm Control Sequences](https://www.xfree86.org/current/ctlseqs.html),
which is designated for clipboard setting. Once a terminal supporting OSC 52 catches a
text in the form of OSC 52 from the output, instead of printing the text onto the screen,
it decodes the text first and then sends the content to the system clipboard.

Although Terminal.app do NOT support OSC 52, here is the workaround for it.

## Installation

Go toolchain is required.

```
GO111MODULE=on go get github.com/roy2220/osc52pty
```

Now you got the binary:

```
ls -lh "$(go env GOPATH)/bin/osc52pty"
```

## Usage

Launch a shell with `osc52pty` to get OSC 52 supported:

```bash
osc52pty bash
```

Within the shell launched, send a OSC 52 sequence to testify:

```bash
printf "\e]52;c;%s\a" "$(echo -n 'THE TEXT TO COPY' | openssl base64 -A)"
```

Now the system clipboard is set to `THE TEXT TO COPY`.

Note: If you're going to send a OSC 52 sequence through TMUX, use this instead:

```bash
printf "\ePtmux;\e\e]52;c;%s\a\e\\" "$(echo -n 'THE TEXT TO COPY' | openssl base64 -A)"
```

BTW, TMUX's clipboard can play well with the OSC 52, search `set-clipboard` in
`man tmux` for more details.

## See also

- [remote-pbcopy-iterm2](https://github.com/skaji/remote-pbcopy-iterm2)
