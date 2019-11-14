package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const restAPIlogName = "restAPI"

type incomeJob struct {
	Some string `json:"some"`
}

// RestAPIstruct restapi entity
type RestAPIstruct struct {
	server        *http.Server
	router        *mux.Router
	programFacade *ProgramFacade
	isUp          chan struct{}
	isDown        chan struct{}
	canRetryUP    chan struct{}
	helpInfo      string
	mainUUID      string
}

// NewRestAPIentity ...
func NewRestAPIentity(ip, port, helpInfo, mainUUID string, programFacade *ProgramFacade) *RestAPIstruct { // TODO: authentication
	router := mux.NewRouter()
	fullAddres := ip + ":" + port
	server := &http.Server{
		Addr: fullAddres, // ip + ":" + port - not working here
		// Good practice to set timeouts to avoid Slowloris attacks.
		// WriteTimeout: time.Second * 15,
		// ReadTimeout:  time.Second * 15,
		// IdleTimeout:  time.Second * 60,
		Handler: router,
	}

	return &RestAPIstruct{
		server:        server,
		router:        router,
		programFacade: programFacade,
		isUp:          make(chan struct{}, 1),
		isDown:        make(chan struct{}, 1),
		canRetryUP:    make(chan struct{}, 1),
		helpInfo:      helpInfo,
		mainUUID:      mainUUID,
	}
}

// UpRestAPI ...
func (restAPI *RestAPIstruct) UpRestAPI(signalChan chan os.Signal, shutdownCommandForRestAPI chan struct{}) {
	restAPI.router.HandleFunc("/", restAPI.rootHandler).Methods("GET")
	restAPI.router.HandleFunc("/dojob", restAPI.doJobRequest).Methods("POST")
	waitRestUpTime := (5 * time.Second)

tryStartRestAPI:
	go restAPI.checkRestAPI(waitRestUpTime)

	if err := restAPI.server.ListenAndServe(); err != http.ErrServerClosed {
		restAPI.isDown <- struct{}{}
		restAPI.programFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": restAPI.mainUUID,
		}).Errorf("rest api down: %v", err)
		select {
		case <-shutdownCommandForRestAPI:
			signalChan <- syscall.SIGTERM
			return
		default:
			<-restAPI.canRetryUP
			time.Sleep(waitRestUpTime)
			goto tryStartRestAPI
		}
	}
}

func (restAPI *RestAPIstruct) doJobRequest(w http.ResponseWriter, r *http.Request) {
	doJobUUID := restAPI.programFacade.UUIDgenerator.NewUUID().UUID.String()

	restAPI.programFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": doJobUUID,
	}).Info("got new doJobRequest request")

	var err error
	buf := new(bytes.Buffer) //read incoming data to buffer, beacose we can't reuse read-closer
	buf.ReadFrom(r.Body)
	bytesFromBuf := buf.Bytes()

	newJob := &incomeJob{}

	err = json.Unmarshal(bytesFromBuf, &newJob)
	if err != nil {
		restAPI.programFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": doJobUUID,
		}).Errorf("can't unmarshal income doJobRequest request: %v", err)

		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !newJob.isThisValidIncomeRequest() {
		restAPI.programFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": doJobUUID,
		}).Error("income doJobRequest request is invalid")

		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = restAPI.programFacade.DoJob(newJob.Some,
		doJobUUID)

	if err != nil {
		restAPI.programFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": doJobUUID,
		}).Errorf("can't doJobRequest, got error: %v", err)

		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}

	restAPI.programFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": doJobUUID,
	}).Info("doJobRequest successfully completed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return

}

func (ig *incomeJob) isThisValidIncomeRequest() bool {
	if ig.Some == "" {
		return false
	}
	return true
}

// GracefulShutdownRestAPI ...
func (restAPI *RestAPIstruct) GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone chan struct{}) {
	<-gracefulShutdownCommandForRestAPI
	restAPI.programFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": restAPI.mainUUID,
	}).Info("stoping http server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Second))
	defer cancel()

	err := restAPI.server.Shutdown(ctx)
	if err != nil {
		restAPI.programFacade.Logging.WithFields(logrus.Fields{
			"entity":     restAPIlogName,
			"event uuid": restAPI.mainUUID,
		}).Errorf("shutdown request error: %v", err)
	}

	restAPI.programFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": restAPI.mainUUID,
	}).Info("rest api stoped")

	restAPIisDone <- struct{}{}
}

func (restAPI *RestAPIstruct) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(restAPI.helpInfo))
}

// checkRestAPI - check rest AIP is up
func (restAPI *RestAPIstruct) checkRestAPI(waitRestUpTime time.Duration) {
	restAPI.programFacade.Logging.WithFields(logrus.Fields{
		"entity":     restAPIlogName,
		"event uuid": restAPI.mainUUID,
	}).Info("check rest api is up")

	addresForGet := "http://" + restAPI.server.Addr + "/"

	ticker := time.NewTicker(time.Duration(100 * time.Millisecond))
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), waitRestUpTime)
	defer cancel()

	checkRestAPIChan := make(chan struct{}, 1)
	checkRestAPIChan <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			restAPI.canRetryUP <- struct{}{}
			return
		case <-checkRestAPIChan:
			resp, err := http.Get(addresForGet)
			if err != nil {
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			if string(body) == restAPI.helpInfo {
				restAPI.programFacade.Logging.WithFields(logrus.Fields{
					"entity":     restAPIlogName,
					"event uuid": restAPI.mainUUID,
				}).Info("rest api is running")
			}
			restAPI.isUp <- struct{}{}
			return
		case <-ticker.C:
			checkRestAPIChan <- struct{}{}
		}
	}
}
