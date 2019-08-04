# Executor

A simple command executor. Supports file paths as arguments.

Incoming data must be an array of bytes **without any encoding**.

## Example

```golang
import "github.com/thevan4/go-billet/executor"
...
fullCommand := "/usr/bin/python"
contextFolder := "/someContextFolder"
arguments := []string{"/opt/some/scripts/template.py", "/execute-file-1464679034681560983"}
stdout, stderr, exitCode, err := Execute(fullCommand, contextFolder, arguments)
if err != nil {
    // your error handling
}
// do something with the stdout/stderr/exitCode
```
