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

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/rammstein4o/git-gpt/utils"
)

const hookFileName = "prepare-commit-msg"

var hookTemplate = `#!/bin/sh

git gpt commit --file $1 --preview
`

var excludeFromDiff = []string{
	"package-lock.json",
	// yarn.lock, Cargo.lock, Gemfile.lock, Pipfile.lock, etc.
	"*.lock",
	"*.snap",
	"go.sum",
}

type Git interface {
	Status() (string, error)
	Commit(val string) (string, error)
	GitDir() (string, error)
	DiffNames() (string, error)
	DiffFile(file string) (string, error)
	ShowDeletedFile(file string) (string, error)
	InstallHook() error
	UninstallHook() error
}

// Ensure, that gitcmd does implement Git.
var _ Git = &gitcmd{}

type gitcmd struct {
	cfg *config
}

func (gc *gitcmd) excludeFiles() []string {
	var excludedFiles []string
	for _, f := range gc.cfg.excludeList {
		excludedFiles = append(excludedFiles, ":(exclude)"+f)
	}
	return excludedFiles
}

func (gc *gitcmd) hookPath() (string, error) {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--git-path",
		"hooks",
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) Status() (string, error) {
	out, err := exec.Command(
		"git",
		"status",
		"--short",
		"--no-renames",
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) Commit(val string) (string, error) {
	out, err := exec.Command(
		"git",
		"commit",
		"--no-verify",
		"--signoff",
		fmt.Sprintf("--message=%s", val),
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

// GitDir to show the (by default, absolute) path of the git directory of the working tree.
func (gc *gitcmd) GitDir() (string, error) {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--git-dir",
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) DiffNames() (string, error) {
	args := []string{
		"diff",
		"--name-only",
		"--staged",
	}

	excludedFiles := gc.excludeFiles()
	args = append(args, excludedFiles...)

	out, err := exec.Command(
		"git",
		args...,
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) DiffFile(file string) (string, error) {
	args := []string{
		"diff",
		"--ignore-all-space",
		"--no-color",
		"--diff-algorithm=minimal",
		fmt.Sprintf("--unified=%d", gc.cfg.diffUnified),
		"--staged",
	}

	excludedFiles := gc.excludeFiles()
	args = append(args, excludedFiles...)
	args = append(args, file)

	out, err := exec.Command(
		"git",
		args...,
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) ShowDeletedFile(file string) (string, error) {
	out, err := exec.Command(
		"git",
		"show",
		fmt.Sprintf("HEAD^:%s", file),
	).Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (gc *gitcmd) InstallHook() error {
	hookPath, err := gc.hookPath()
	if err != nil {
		return err
	}

	target := path.Join(strings.TrimSpace(hookPath), hookFileName)
	if utils.IsFile(target) {
		return errors.New("hook file prepare-commit-msg exist")
	}

	content, err := utils.GetTemplateByBytes(hookTemplate, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(target, content, 0o755)
}

func (gc *gitcmd) UninstallHook() error {
	hookPath, err := gc.hookPath()
	if err != nil {
		return err
	}

	target := path.Join(strings.TrimSpace(hookPath), hookFileName)
	if !utils.IsFile(target) {
		return errors.New("hook file prepare-commit-msg does not exist")
	}
	return os.Remove(target)
}

func New(opts ...Option) Git {
	// Instantiate a new config object with default values
	cfg := &config{}

	// Loop through each option passed as argument and apply it to the config object
	for _, fn := range opts {
		fn(cfg)
	}

	// Append the user-defined excludeList to the default excludeFromDiff
	cfg.excludeList = append(excludeFromDiff, cfg.excludeList...)

	return &gitcmd{
		cfg: cfg,
	}
}
