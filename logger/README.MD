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
package main

import (
	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/logger"
)

func main() {
	newLogger := &logger.Logger{
		Output:           []string{"stdout"},
		Level:            "debug",
		Formatter:        "text",
		LogEventLocation: true,
	}
	logrusLog, err := logger.NewLogrusLogger(newLogger)
	if err != nil {
		panic(err)
	}

	logrusLog.WithFields(logrus.Fields{
		"entity":     "main",
		"event uuid": "00000000-0000-0000-0000-000000000000",
	}).Info("Hi, i'am logger")
}
```

```golang
package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/thevan4/go-billet/logger"
)

func main() {
	client := &http.Client{}
	caCert, err := ioutil.ReadFile("./cert/server.crt")
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig
	}
	httpClient = &http.Client{Transport: httpTransport}
	

	newGraylog := &logger.Graylog{
		Address: "http://graylog.sdc.com:12201/gelf",
		Retries: 30,
		HTTPClient: httpClient,
	}
	newLogger := &logger.Logger{
		Output:    []string{"stdout"},
		Level:     "info",
		Formatter: "json",
		Graylog:   newGraylog,
	}
	logrusLog, err := logger.NewLogrusLogger(newLogger)
	if err != nil {
		panic(err)
	}

	logrusLog.WithFields(logrus.Fields{
		"entity":     "main",
		"event uuid": "00000000-0000-0000-0000-000000000000",
	}).Info("Hi, i'am logger")
}
```
