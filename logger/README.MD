# Logger

Used to create a logging entity.

Libraries used:
[sirupsen: logrus](github.com/sirupsen/logrus)
[graylog-http-hook](github.com/thevan4/logrus-graylog-http-hook)

## Start working

To start working with logging you need to initialize the "Logger" structure with the necessary parameters:

+ Logger.Output - stdout, syslog, graylog.
+ Logger.Level - info, debug, trace.
+ Logger.Formatter - json, text.
+ Logger.SyslogTag - any string (if necessary).
+ Logger.Graylog - [instructions below](#Graylog-settings).
+ Logger.LogEventLocation - Logging the event location: the file, the stack, and the name of the function where the logger was called.

### <a name="Graylog-settings"></a>Graylog settings

For work with graylog you need to initialize the "Graylog" structure:

+ Graylog.Address - address format: "http://graylog.sdc.com:12201/gelf".
+ Graylog.Retries - repeated attempts to send if the connection to graylog failed
+ Graylog.Extra - some extra info, format  map[string]interface{}.

## Examples

```golang

```

```golang

```
