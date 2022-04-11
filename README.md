<!-- ⚠️ This README has been generated from the file(s) ".config/docs/blueprint-readme-cli.md" ⚠️--><div align="center">
  <center>
    <a href="https://github.com/megabyte-labs/Bodega">
      <img width="148" height="148" alt="Bodega logo" src="https://gitlab.com/megabyte-labs/go/cli/bodega/-/raw/master/logo.png" />
    </a>
  </center>
</div>
<div align="center">
  <center><h1 align="center"><i></i>Bodega - Your Go-To Task Runner<i></i></h1></center>
  <center><h4 style="color: #18c3d1;"><i></i><a href="https://megabyte.space" target="_blank">Megabyte Labs</a></h4><i></i></center>
</div>

<div align="center">
  <a href="https://megabyte.space" title="Megabyte Labs homepage" target="_blank">
    <img alt="Homepage" src="https://img.shields.io/website?down_color=%23FF4136&down_message=Down&label=Homepage&logo=home-assistant&logoColor=white&up_color=%232ECC40&up_message=Up&url=https%3A%2F%2Fmegabyte.space&style=for-the-badge" />
  </a>
  <a href="https://github.com/megabyte-labs/Bodega/blob/master/docs/CONTRIBUTING.md" title="Learn about contributing" target="_blank">
    <img alt="Contributing" src="https://img.shields.io/badge/Contributing-Guide-0074D9?logo=github-sponsors&logoColor=white&style=for-the-badge" />
  </a>
  <a href="https://app.slack.com/client/T01ABCG4NK1/C01NN74H0LW/details/" title="Chat with us on Slack" target="_blank">
    <img alt="Slack" src="https://img.shields.io/badge/Slack-Chat-e01e5a?logo=slack&logoColor=white&style=for-the-badge" />
  </a>
  <a href="link.gitter" title="Chat with the community on Gitter" target="_blank">
    <img alt="Gitter" src="https://img.shields.io/gitter/room/megabyte-labs/community?logo=gitter&logoColor=white&style=for-the-badge" />
  </a>
  <a href="https://github.com/megabyte-labs/Bodega" title="GitHub mirror" target="_blank">
    <img alt="GitHub" src="https://img.shields.io/badge/Mirror-GitHub-333333?logo=github&style=for-the-badge" />
  </a>
  <a href="https://gitlab.com/megabyte-labs/go/cli/bodega" title="GitLab repository" target="_blank">
    <img alt="GitLab" src="https://img.shields.io/badge/Repo-GitLab-fc6d26?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgAQMAAABJtOi3AAAABlBMVEUAAAD///+l2Z/dAAAAAXRSTlMAQObYZgAAAHJJREFUCNdNxKENwzAQQNEfWU1ZPUF1cxR5lYxQqQMkLEsUdIxCM7PMkMgLGB6wopxkYvAeI0xdHkqXgCLL0Beiqy2CmUIdeYs+WioqVF9C6/RlZvblRNZD8etRuKe843KKkBPw2azX13r+rdvPctEaFi4NVzAN2FhJMQAAAABJRU5ErkJggg==&style=for-the-badge" />
  </a>
</div>
<br/>
<div align="center">
  <a title="Version: 11.0.0" href="https://github.com/megabyte-labs/Bodega" target="_blank">
    <img alt="Version: 11.0.0" src="https://img.shields.io/badge/version-11.0.0-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgAQMAAABJtOi3AAAABlBMVEUAAAD///+l2Z/dAAAAAXRSTlMAQObYZgAAACNJREFUCNdjIACY//+BEp9hhM3hAzYQwoBIAqEDYQrCZLwAAGlFKxU1nF9cAAAAAElFTkSuQmCC&cacheSeconds=2592000&style=flat-square" />
  </a>
  <a title="Go version: goVersion" href="https://github.com/megabyte-labs/Bodega/blob/master/go.mod" target="_blank">
    <img alt="Go version: goVersion" src="https://img.shields.io/github/go-mod/go-version/profile.github}}/{{slug?logo=go&logoColor=white&style=flat-square">
  </a>
  <a title="GitLab build status" href="https://gitlab.com/megabyte-labs/go/cli/bodega/-/commits/master" target="_blank">
    <img alt="Build status" src="https://img.shields.io/gitlab/pipeline-status/megabyte-labs/ansible-roles/galaxy_info.role_name?branch=master&label=build&logo=gitlab&style=flat-square">
  </a>
  <a title="Documentation" href="https://megabyte.space/docs/go" target="_blank">
    <img alt="Documentation" src="https://img.shields.io/badge/documentation-yes-brightgreen.svg?logo=readthedocs&style=flat-square" />
  </a>
  <a title="License: MIT" href="https://github.com/megabyte-labs/Bodega/blob/master/LICENSE" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-yellow.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgAQMAAABJtOi3AAAABlBMVEUAAAD///+l2Z/dAAAAAXRSTlMAQObYZgAAAHpJREFUCNdjYOD/wMDAUP+PgYHxhzwDA/MB5gMM7AwMDxj4GBgKGGQYGCyAEEgbMDDwAAWAwmk8958xpIOI5zKH2RmOyhxmZjguAiKmgIgtQOIYmFgCIp4AlaQ9OczGkJYCJEAGgI0CGwo2HmwR2Eqw5SBnNIAdBHYaAJb6KLM15W/CAAAAAElFTkSuQmCC&style=flat-square" />
  </a>
</div>

> </br><h4 align="center">**A gorgeous, feature-packed drop-in replacement for the [go-task/task](https://github.com/go-task/task) project**</h4></br>

<!--TERMINALIZE![terminalizer_title](https://gitlab.com/megabyte-labs/go/cli/bodega/-/raw/master/docs/demo.gif)TERMINALIZE-->

<a href="#table-of-contents" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## Table of Contents

- [Overview](#overview)
- [This Repository (Shared Common)](#this-repository-shared-common)
- [Installation](#installation)
  - [Quick Method](#quick-method)
  - [Compile Program with Go](#compile-program-with-go)
  - [NPM Install Method](#npm-install-method)
  - [Pre-Built Binary](#pre-built-binary)
- [Usage](#usage)
  - [Man Page](#man-page)
- [Contributing](#contributing)
- [License](#license)

<a href="#overview" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## Overview

**Bodega** is a fork of [go-task](https://github.com/go-task/task) that brings a slew of new features. It adds powerful UX features and functional improvements while retaining backwards compatibility. For those of you who do not know, Task is a task runner / Make alternative written in Go. It allows you to define bash script snippets in YML files and provides some advanced features. Some of the features it includes is the ability to define dependencies, running tasks conditionally, caching values, and a built-in method of housing CLI documentation. **Bodega** takes go-task to the next level by improving the TUI experience and including features that allow you to use the project as a CLI-generator.

<a href="#this-repository-shared-common" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## This Repository (Shared Common)

This repository houses files that propagate downstream to project-specific repositories.

<a href="#installation" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## Installation

There are several ways you can install this CLI. You can:

1. Use our bash scripts which will handle everything automatically with as few dependencies as possible
2. Compile the program using Go and add it to your `PATH`
3. Install it via an NPM convienience wrapper
4. Download the pre-built binary from the GitLab or GitHub releases page and then place it in your `PATH`

### Quick Method

If you are looking to install the CLI as quickly as possible then you can run the following script which will install the binary to your `/usr/local/bin` folder on macOS or Linux:

```
curl -sS https://install.doctor/task | bash
```

Or, if you are on Windows, you can install it by running:

```
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://install.doctor/task?os=win'))
```

### Compile Program with Go

You can install the CLI by compiling it from the source as long as you have a recent version of Go installed:

```
git clone https://github.com/megabyte-labs/Bodega.git
cd {{#withLast (split repository.github "/")}}this}}{{/withLast
go build -o dist/task cmd/task/task.go
sudo mv ./dist/task /usr/local/bin
```

After you compile the program, you should then move the binary file to a location that is in your `PATH` (which is what the last line does in the snippet above).

### NPM Install Method

Every release is bundled into an NPM package that you can install by running the following command:

```
npm install -g task
```

### Pre-Built Binary

If you trust us (and you should not.. trust.. anybody.. EVER), then you can also download the binary directly from the Bodega GitLab release page or the GitHub release page. After you download the release, you will have to either place the binary somewhere in your `PATH` or run the installer (in the case of the `.deb` or `.rpm` releases, for instance).

<a href="#usage" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## Usage

All of the usage instructions can be found by running `task --help`. After running the command, you should be greeted with the following output:

```
Usage: task [-ilfwvsdm] [--init] [--list] [--force] [--watch] [--verbose] [--silent] [--dir] [--taskfile] [--dry] [--menu] [--summary] [--debug] [task...]

Runs the specified task(s). Runs a built-in shell if no task name
was specified, or lists all tasks if an unknown task name was specified.

Example: 'task hello' with the following 'Taskfile.yml' file will generate an
'output.txt' file with the content "hello".

'''
version: '3'
tasks:
  hello:
    cmds:
      - echo "I am going to write a file named 'output.txt' now."
      - echo "hello" > output.txt
    generates:
      - output.txt
'''

Options:
  -c, --color             colored output. Enabled by default. Set flag to false or use NO_COLOR=1 to disable (default true)
  -C, --concurrency int   limit number tasks to run concurrently
      --debug             stop before each command execution
  -d, --dir string        sets directory of execution
      --dry               compiles and prints tasks in the order that they would be run, without executing them
  -f, --force             forces execution even when the task is up-to-date
  -h, --help              shows Task usage
  -i, --init              creates a new Taskfile.yaml in the current folder
  -l, --list              lists tasks with description of current Taskfile
  -a, --list-all          lists tasks with or without a description
  -m, --menu              runs an interactive listing of tasks
  -o, --output string     sets output style: [interleaved|group|prefixed]
  -p, --parallel          executes tasks provided on command line in parallel
      --server            runs as a server
  -s, --silent            disables echoing
      --status            exits with non-zero exit code if any of the given tasks is not up-to-date
      --summary           show summary about a task
  -t, --taskfile string   choose which Taskfile to run. Defaults to "Taskfile.yml"
      --use-tls           enable server to use TLS
  -v, --verbose count     enables verbose mode (repeat option for more output)
      --version           show Task version
  -w, --watch             enables watch of the given task
```

### Man Page

Alternatively, if you installed the package via NPM or an installer that set up the man page (e.g. `.deb` or `.rpm`), then you can find usage instructions by running `man task`.

<a href="#contributing" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/megabyte-labs/Bodega/issues). If you would like to contribute, please take a look at the [contributing guide](https://github.com/megabyte-labs/Bodega/blob/master/docs/CONTRIBUTING.md).

<details>
<summary><b>Sponsorship</b></summary>
<br/>
<blockquote>
<br/>
Dear Awesome Person,<br/><br/>
I create open source projects out of love. Although I have a job, shelter, and as much fast food as I can handle, it would still be pretty cool to be appreciated by the community for something I have spent a lot of time and money on. Please consider sponsoring me! Who knows? Maybe I will be able to quit my job and publish open source full time.
<br/><br/>Sincerely,<br/><br/>

**_Brian Zalewski_**<br/><br/>

</blockquote>

<a title="Support us on Open Collective" href="https://opencollective.com/megabytelabs" target="_blank">
  <img alt="Open Collective sponsors" src="https://img.shields.io/opencollective/sponsors/megabytelabs?logo=opencollective&label=OpenCollective&logoColor=white&style=for-the-badge" />
</a>
<a title="Support us on GitHub" href="https://github.com/ProfessorManhattan" target="_blank">
  <img alt="GitHub sponsors" src="https://img.shields.io/github/sponsors/ProfessorManhattan?label=GitHub%20sponsors&logo=github&style=for-the-badge" />
</a>
<a href="https://www.patreon.com/ProfessorManhattan" title="Support us on Patreon" target="_blank">
  <img alt="Patreon" src="https://img.shields.io/badge/Patreon-Support-052d49?logo=patreon&logoColor=white&style=for-the-badge" />
</a>

</details>

<a href="#license" style="width:100%"><img style="width:100%" src="https://gitlab.com/megabyte-labs/assets/-/raw/master/png/aqua-divider.png" /></a>

## License

Copyright © 2020-2021 [Megabyte LLC](https://megabyte.space). This project is [MIT](https://gitlab.com/megabyte-labs/go/cli/bodega/-/blob/master/LICENSE) licensed.
