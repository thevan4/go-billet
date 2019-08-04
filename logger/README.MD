# Logger

Used to create a logging entity in the program.

Libraries used:
[sirupsen: logrus](github.com/sirupsen/logrus)

Possible values:

**logOutput** - stdout, syslog.

**logLevel** - info, debug, trace.

**logFormat** - json, default.

**syslogTag** - any string.

## Examples

```golang
import "github.com/thevan4/go-billet/logger"
...
logging, err := logger.NewLogger(logOutput, logLevel, logFormat, syslogTag)
if err != nil {
    // your error handling
}
logging.WithFields(logrus.Fields{
    "entity":     someEntity,
    "event uuid": "00000000-0000-0000-0000-000000000000",
    "message":    someMessage,
}).Info("Hi, i'am logger")
```

```golang
import "github.com/thevan4/go-billet/logger"
...
logging, err := logger.NewLogger("stdout", "info", "default", "")
if err != nil {
    fmt.Println(err)
    os.Exit(1)
}
```
