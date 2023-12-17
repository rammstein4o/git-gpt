// ---
// Copyright © 2023 Radoslav Salov <rado.salov@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
// ---

package git

import "strings"

type GitOperation string

const (
	OPERATION_ADD GitOperation = "A"
	OPERATION_DEL GitOperation = "D"
	OPERATION_MOD GitOperation = "M"
)

func ParseGitStatus(status string) (added, removed, modified []string) {
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := GitOperation(strings.TrimSpace(line[:2]))
		file := strings.TrimSpace(line[3:])

		switch status {
		case OPERATION_ADD:
			added = append(added, file)
		case OPERATION_DEL:
			removed = append(removed, file)
		case OPERATION_MOD:
			modified = append(modified, file)
		}
	}

	return added, removed, modified
}
