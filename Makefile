###
# Copyright © 2023 Radoslav Salov <rado.salov@gmail.com>
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
###
.PHONY: all
all: clean build

.PHONY: clean
clean:
	rm -rf ./dist

.PHONY: build-linux
build-linux:
	mkdir -p ./dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -a -o ./dist/git-gpt-linux-amd64 github.com/rammstein4o/git-gpt

.PHONY: build-windows
build-windows:
	mkdir -p ./dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -a -o ./dist/git-gpt-win-amd64.exe github.com/rammstein4o/git-gpt

.PHONY: build-macos
build-macos:
	mkdir -p ./dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -a -o ./dist/git-gpt-darwin-amd64 github.com/rammstein4o/git-gpt

.PHONY: build
build: clean build-linux build-windows build-macos
