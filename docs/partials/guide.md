### Run Task as a Server
https://gitlab.com/megabyte-labs/go/cli/bodega/-/commit/80a622abcbe54d693216b048902dc3e0cf3ef92a updates the API to include a basic run command. You can test it live by spawning it with `--server` flag and sending websocket JSON messages. Here's an example format:
```json
{
  "command": "run",
  "tasks": ["clean", "simple"],
  "options": {
    # verbose levels are 0 (none), 1 (normal) and 2 (debug)
    "verbose": 1,
    "force": false,
    "silent": false,
    "nLines": 5
  }
}
```
Currently "run" and "list" commands are supported, and the options mentioned above. Here's an example for "list" command:
```json
{
    "command": "list"
}
```

parameter `nLines` is used to control the number of output lines you receive until the task finishes. Connection is closed for each run command with:

* Status code of 3002 for failed runs with message: "running task failed"
* Status code of 3001 for successful runs with message: "task exited successfully"

If you would like to simulate secure communication with the `--use-tls` flag, you could Issue a signed certificate and a private key to use and name them "localhost". You can download [mkcert](https://github.com/FiloSottile/mkcert) and issue the command `mkcert localhost`. This will create a private key and a self-signed certificate named localhost. Put them in Task directory and it should read them. This should change in the future to read from a file

### Run task once per Task invocation
You can run a task once per invocation of the Task program. Running a task once will persist a temporary file in user's home directory. Here's an example task:

```yaml
version: '3'

run: once
run_once_system: true

tasks:
  install-deps:
    run: once
    run_once_system: false
    cmds:
      - sleep 5 # long operation like installing packages
  install-deps-system-wide:
    run: once
    # run_once_system: true # already set at the task level
    cmds:
      - sleep 10
```

Please note that `run_once_system` is a boolean variable that only works when `run` value of `once` is specified. Its default value is false.

### Stop commands before execution
Passing the `--debug` command-line flag makes Task stop before each command execution, even for commands within a variable.

```
$ ./task --debug simple
task: [simple] echo 'hi'
Executing a shell command. Type enter to continue

hi

```
