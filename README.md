## gowipe - securely delete files and directories (not for SSD)

[![Actions](https://github.com/tlinden/gowipe/actions/workflows/ci.yaml/badge.svg)](https://github.com/tlinden/gowipe/actions)
[![License](https://img.shields.io/badge/license-GPL-blue.svg)](https://github.com/tlinden/gowipe/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/tlinden/gowipe)](https://goreportcard.com/report/github.com/tlinden/gowipe)
[![GitHub release](https://img.shields.io/github/v/release/tlinden/gowipe?color=%2300a719)](https://github.com/TLINDEN/gowipe/releases/latest)


```
Usage: gowipe [-rcvz] <file|directory>...

Options:
-r --recursive    Delete <dir> recursively
-c --count <num>  Overwrite files <num> times
-m --mode <mode>  Use <mode> for overwriting (or use -E, -S, -M, -Z)
-n --nodelete     Do not delete files after overwriting
-N --norename     Do not rename the files
-v --verbose      Verbose output
-V --version      Show program version
-h --help         Show usage

Available modes:
zero      Overwrite with zeroes (-Z)
math      Overwrite with math random bytes (-M)
secure    Overwrite with secure random bytes (default) (-S)
encrypt   Overwrite with ChaCha2Poly1305 encryption (most secure) (-E)
```

## Getting help

Although I'm happy to hear from gowipe users in private email,
that's the best way for me to forget to do something.

In order to report a bug, unexpected behavior, feature requests
or to submit a patch, please open an issue on github:
https://github.com/TLINDEN/gowipe/issues.

## Copyright and license

This software is licensed under the GNU GENERAL PUBLIC LICENSE version 3.

## Authors

T.v.Dein <tom AT vondein DOT org>

## Project homepage

https://github.com/TLINDEN/gowipe
