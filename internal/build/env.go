// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package build

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	// These flags override values in build env.
	GitCommitFlag   = flag.String("git-commit", "", `Overrides git commit hash embedded into executables`)
	GitBranchFlag   = flag.String("git-branch", "", `Overrides git branch being built`)
	GitTagFlag      = flag.String("git-tag", "", `Overrides git tag being built`)
	BuildnumFlag    = flag.String("buildnum", "", `Overrides CI build number`)
	PullRequestFlag = flag.Bool("pull-request", false, `Overrides pull request status of the build`)
	CronJobFlag     = flag.Bool("cron-job", false, `Overrides cron job status of the build`)
)

// Environment contains metadata provided by the build environment.
type Environment struct {
	Name                string // name of the environment
	Repo                string // name of GitHub repo
	Commit, Branch, Tag string // Git info
	Buildnum            string
	IsPullRequest       bool
	IsCronJob           bool
}

func (env Environment) String() string {
	return fmt.Sprintf("%s env (commit:%s branch:%s tag:%s buildnum:%s pr:%t)",
		env.Name, env.Commit, env.Branch, env.Tag, env.Buildnum, env.IsPullRequest)
}

// Env returns metadata about the current CI environment, falling back to LocalEnv
// if not running on CI.
func Env() Environment {
	switch {
	case os.Getenv("CI") == "true" && os.Getenv("TRAVIS") == "true":
		return Environment{
			Name:          "travis",
			Repo:          os.Getenv("TRAVIS_REPO_SLUG"),
			Commit:        os.Getenv("TRAVIS_COMMIT"),
			Branch:        os.Getenv("TRAVIS_BRANCH"),
			Tag:           os.Getenv("TRAVIS_TAG"),
			Buildnum:      os.Getenv("TRAVIS_BUILD_NUMBER"),
			IsPullRequest: os.Getenv("TRAVIS_PULL_REQUEST") != "false",
			IsCronJob:     os.Getenv("TRAVIS_EVENT_TYPE") == "cron",
		}
	case os.Getenv("CI") == "True" && os.Getenv("APPVEYOR") == "True":
		return Environment{
			Name:          "appveyor",
			Repo:          os.Getenv("APPVEYOR_REPO_NAME"),
			Commit:        os.Getenv("APPVEYOR_REPO_COMMIT"),
			Branch:        os.Getenv("APPVEYOR_REPO_BRANCH"),
			Tag:           os.Getenv("APPVEYOR_REPO_TAG_NAME"),
			Buildnum:      os.Getenv("APPVEYOR_BUILD_NUMBER"),
			IsPullRequest: os.Getenv("APPVEYOR_PULL_REQUEST_NUMBER") != "",
			IsCronJob:     os.Getenv("APPVEYOR_SCHEDULED_BUILD") == "True",
		}
	default:
		return LocalEnv()
	}
}

// LocalEnv returns build environment metadata gathered from git.
func LocalEnv() Environment {
	env := applyEnvFlags(Environment{Name: "local", Repo: "matrix/go-matrix"})

	head := readGitFile("HEAD")
	if splits := strings.Split(head, " "); len(splits) == 2 {
		head = splits[1]
	} else {
		return env
	}
	if env.Commit == "" {
		env.Commit = readGitFile(head)
	}
	if env.Branch == "" {
		if head != "HEAD" {
			env.Branch = strings.TrimPrefix(head, "refs/heads/")
		}
	}
	if info, err := os.Stat(".git/objects"); err == nil && info.IsDir() && env.Tag == "" {
		env.Tag = firstLine(RunGit("tag", "-l", "--points-at", "HEAD"))
	}
	return env
}

func firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

func applyEnvFlags(env Environment) Environment {
	if !flag.Parsed() {
		panic("you need to call flag.Parse before Env or LocalEnv")
	}
	if *GitCommitFlag != "" {
		env.Commit = *GitCommitFlag
	}
	if *GitBranchFlag != "" {
		env.Branch = *GitBranchFlag
	}
	if *GitTagFlag != "" {
		env.Tag = *GitTagFlag
	}
	if *BuildnumFlag != "" {
		env.Buildnum = *BuildnumFlag
	}
	if *PullRequestFlag {
		env.IsPullRequest = true
	}
	if *CronJobFlag {
		env.IsCronJob = true
	}
	return env
}
