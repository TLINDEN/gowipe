# Copyright © 2023 Thomas von Dein

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.


#
# no need to modify anything below
tool      = gowipe
VERSION   = $(shell grep VERSION main.go | head -1 | cut -d '"' -f2)
archs     = darwin freebsd linux windows
PREFIX    = /usr/local
UID       = root
GID       = 0
HAVE_POD := 

all: $(tool) buildlocal

buildlocal:
	CGO_LDFLAGS='-static' go build -tags osusergo,netgo -ldflags "-extldflags=-static" -o $(tool)

install: buildlocal
	install -d -o $(UID) -g $(GID) $(PREFIX)/bin
	install -d -o $(UID) -g $(GID) $(PREFIX)/man/man1
	install -o $(UID) -g $(GID) -m 555 $(tool)  $(PREFIX)/sbin/
	install -o $(UID) -g $(GID) -m 444 $(tool).1 $(PREFIX)/man/man1/

clean:
	rm -rf $(tool) coverage.out

test:
	go test -v ./...

singletest:
	@echo "Call like this: ''make singletest TEST=TestPrepareColumns"
	go test -run $(TEST)

cover-report:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

goupdate:
	go get -t -u=patch ./...

buildall:
	./mkrel.sh $(tool) $(VERSION)

release: buildall
	gh release create v$(VERSION) --generate-notes releases/*

show-versions: buildlocal
	@echo "### gowipe version:"
	@./gowipe -v

	@echo
	@echo "### go module versions:"
	@go list -m all

	@echo
	@echo "### go version used for building:"
	@grep -m 1 go go.mod


dir:
	rm -rf a
	mkdir -p a/b/c
	date > a/filea
	date > a/b/fileb
	date > a/b/c/filec

bench: all
	dd if=/dev/zero of=t/fileZ bs=1024 count=200000
	dd if=/dev/zero of=t/fileM bs=1024 count=200000
	dd if=/dev/zero of=t/fileS bs=1024 count=200000
	dd if=/dev/zero of=t/fileE bs=1024 count=200000
	/usr/bin/time -f "%S" ./gowipe -Z t/fileZ
	/usr/bin/time -f "%S" ./gowipe -M t/fileM
	/usr/bin/time -f "%S" ./gowipe -S t/fileS
	/usr/bin/time -f "%S" ./gowipe -E t/fileE
