package portadapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"../domain"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	consumerCluster "github.com/bsm/sarama-cluster"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const consumerLogPrefix = "sarama.consumer"

var consumerConfig *consumerCluster.Config

func init() {
	consumerConfig = consumerCluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	consumerConfig.Version = sarama.V2_1_0_0
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
}

// Consumer ...
type Consumer struct {
	consumerClient *consumerCluster.Consumer
	isUp           chan struct{}
	isDown         chan struct{}
	logging        *logrus.Logger
}

func newConsumer(bootstrapServers, topics []string, groupID string, uuidForNewSaramaPortAdapter string, monitoring domain.PrometheusConfig, logging *logrus.Logger) (*Consumer, error) {
	consumerConfig.Net.ReadTimeout = time.Duration(30 * time.Second)
	client, err := newConsumerClient(bootstrapServers)
	if err != nil {
		return nil, err
	}

	consumerFromClient, err := newConsumerFromClient(client, topics, groupID)
	if err != nil {
		return nil, err
	}

	logging.WithFields(logrus.Fields{
		"entity":     consumerLogPrefix,
		"event uuid": uuidForNewSaramaPortAdapter,
	}).Info("successfully create consumer client")

	isUp, isDown := monitoring.NewHealthMetric("program_sarama_health", "sarama", uuidForNewSaramaPortAdapter)

	consumer := &Consumer{
		consumerClient: consumerFromClient,
		isUp:           isUp,
		isDown:         isDown,
		logging:        logging,
	}

	return consumer, nil
}

// ConnectConsumer ...
func (consumer *Consumer) ConnectConsumer(topics []string, uuidForConnectConsumer string) error {
	consumer.logging.WithFields(logrus.Fields{
		"entity":     consumerLogPrefix,
		"event uuid": uuidForConnectConsumer,
	}).Info("try to connect consumer client")
	claimed := make(chan struct{}, 1)
	go func() {
		for ntf := range consumer.consumerClient.Notifications() {
			if ntf.Type == cluster.RebalanceOK {
				if len(ntf.Claimed) == 0 {
					break
				}

				for _, t := range topics {
					if _, ok := ntf.Claimed[t]; !ok {
						break
					}
				}

				claimed <- struct{}{}
			}
		}
	}()
	select { // TODO: select shutdown signal saramaPortAdapter.shutdownCommandForListen:
	case <-claimed:
		consumer.logging.WithFields(logrus.Fields{
			"entity":     consumerLogPrefix,
			"event uuid": uuidForConnectConsumer,
		}).Info("got rebalance, successfully connect consumer")
		consumer.isUp <- struct{}{} // for first up
		//
		go func() {
			for ntf := range consumer.consumerClient.Notifications() {
				if ntf.Type == cluster.RebalanceOK {
					consumer.logging.WithFields(logrus.Fields{
						"entity":     consumerLogPrefix,
						"event uuid": uuidForConnectConsumer,
					}).Infof("sarama rebalanced: %v", ntf)
					consumer.isUp <- struct{}{}
				}
			}
		}()

		go func() {
			for err := range consumer.consumerClient.Errors() {
				consumer.logging.WithFields(logrus.Fields{
					"entity":     consumerLogPrefix,
					"event uuid": uuidForConnectConsumer,
				}).Errorf("sarama got error: %v", err)
				if strings.Contains(err.Error(), "client has run out of available brokers to talk to") ||
					strings.Contains(err.Error(), "Is your cluster reachable?") ||
					strings.Contains(err.Error(), "i/o timeout") {
					consumer.logging.WithFields(logrus.Fields{
						"entity":     consumerLogPrefix,
						"event uuid": uuidForConnectConsumer,
					}).Errorf("consumer can't read message: %v", err)
					consumer.isDown <- struct{}{}
				}
			}
		}()

		//
		return nil
	case <-time.After(time.Duration(60 * time.Second)):
		return fmt.Errorf("consumer rebalance catch timeout, unable to connect in %s seconds", time.Duration(60*time.Second))
	}
}

// ListenAndHandle ...
func (saramaPortAdapter *SaramaPortAdapter) ListenAndHandle() {
	for {
		select {
		case <-saramaPortAdapter.shutdownCommandForListen:
			saramaPortAdapter.Consumer.consumerClient.Close()
		case consumerMessage := <-saramaPortAdapter.Consumer.consumerClient.Messages():
			// when sarama close chan consumerClient.Messages() we get a lot of null pointers messages
			if consumerMessage != nil {
				saramaPortAdapter.Consumer.isUp <- struct{}{}
				tempNewJobUUID := saramaPortAdapter.uuidgenerator.NewUUID().UUID.String()
				// in some cases we may not have time to process the message and addNewJob. That fix that
				saramaPortAdapter.addNewJob(tempNewJobUUID)
				mapToJobCham := make(map[string]*sarama.ConsumerMessage)
				mapToJobCham[tempNewJobUUID] = consumerMessage
				saramaPortAdapter.consumerJobChanel <- mapToJobCham
			}
		}
	}
}

// consumer handler handle messages from buffered consumer job channel.
func (saramaPortAdapter *SaramaPortAdapter) consumerHandler() {
	for tempNewJobUUID, consumerMessage := range <-saramaPortAdapter.consumerJobChanel {
		go saramaPortAdapter.workWithIncomingMessage(consumerMessage, tempNewJobUUID)
	}
}

func (saramaPortAdapter *SaramaPortAdapter) workWithIncomingMessage(consumerMessage *sarama.ConsumerMessage,
	tempNewJobUUID string) {
	saramaPortAdapter.Consumer.consumerClient.MarkOffset(consumerMessage, "") // TODO: mark message only when in for this program instanse
	// if thatForMe(msg, saramaPortAdapter.ProgramStruct.Some) {
	// 	continue
	// }

	msg, err := unmarshalIncome(consumerMessage)
	if err != nil { // TODO: never return error. need some dirty valdates.. report back
		// failmsg := domain.Message send
		saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
			"entity":     consumerLogPrefix,
			"event uuid": tempNewJobUUID,
		}).Errorf("can't unmarshal new message. Got error: %v", err)
		return
	}

	msg.InternalUUID = msg.MessageID.UUID.String()
	saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
		"entity":     consumerLogPrefix,
		"event uuid": msg.InternalUUID,
	}).Infof("got new message, type: %v Message type: ", msg.Type)

	handler, ok := saramaPortAdapter.HandlerList[msg.Type]

	if !ok {
		saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
			"entity":     consumerLogPrefix,
			"event uuid": msg.InternalUUID,
		}).Errorf("unknown message type %v, can't hande that", msg.Type)
		// TODO: response for that
		return
	}
	saramaPortAdapter.addNewJob(msg.InternalUUID)
	saramaPortAdapter.removeDoneJob(tempNewJobUUID)
	handler.ExecuteHandler(*msg)
	saramaPortAdapter.removeDoneJob(msg.InternalUUID)
}

// GracefulShutdownSaramaJobs ...
func (saramaPortAdapter *SaramaPortAdapter) GracefulShutdownSaramaJobs(shutdownCommandForSaramaListenAndHandle, saramaListenAndHandleIsDone chan struct{}) {
	<-shutdownCommandForSaramaListenAndHandle
	saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
		"entity":     consumerLogPrefix,
		"event uuid": "00000000-0000-0000-0000-000000000000", // TODO: no zero uuids
	}).Info("stoping sarama entity")

	saramaPortAdapter.shutdownCommandForListen <- struct{}{} // dont recive new messages

	time.Sleep(100 * time.Millisecond) // to be sure - all messages from closed chan are readed

	// check all handled works is stoped
	saramaPortAdapter.jobWatcher(time.Duration(30*time.Second), time.Duration(500*time.Millisecond)) // watch at job map

	saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
		"entity":     consumerLogPrefix,
		"event uuid": "00000000-0000-0000-0000-000000000000", // TODO: no zero uuids
	}).Info("sarama entity stopped")
	saramaListenAndHandleIsDone <- struct{}{}
}

func (saramaPortAdapter *SaramaPortAdapter) jobWatcher(timeWaitForStop, updateJobPollInterval time.Duration) {
	ticker := time.NewTicker(updateJobPollInterval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), timeWaitForStop)
	defer cancel()

	checkJobs := make(chan struct{}, 1)
	checkJobs <- struct{}{}

	saramaPortAdapter.jobs.rmx.Lock()
	if len(saramaPortAdapter.jobs.pool) != 0 {
		saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
			"entity":     consumerLogPrefix,
			"event uuid": "00000000-0000-0000-0000-000000000000", // TODO: no zero uuids
		}).Warnf("still has %v jobs in progress when got stop signal", len(saramaPortAdapter.jobs.pool))
	}
	saramaPortAdapter.jobs.rmx.Unlock()

	for {
		select {
		case <-checkJobs:
			saramaPortAdapter.jobs.rmx.Lock()
			if len(saramaPortAdapter.jobs.pool) == 0 {
				saramaPortAdapter.jobs.rmx.Unlock()
				return
			}
			saramaPortAdapter.jobs.rmx.Unlock()
		case <-ctx.Done():
			for jobUUID := range saramaPortAdapter.jobs.pool {
				saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
					"entity":     consumerLogPrefix,
					"event uuid": jobUUID,
				}).Error("task failed due to timeout")
			}
			return
		case <-ticker.C:
			checkJobs <- struct{}{}
		}
	}
}

func newConsumerClient(bootstrapServers []string) (*consumerCluster.Client, error) {
	consumerClient, err := consumerCluster.NewClient(bootstrapServers, consumerConfig)
	if err != nil {
		return nil, err
	}
	return consumerClient, nil
}

func newConsumerFromClient(consumerClient *consumerCluster.Client, topics []string, groupID string) (*consumerCluster.Consumer, error) {
	consumerFromClient, err := consumerCluster.NewConsumerFromClient(consumerClient, groupID, topics)
	if err != nil {
		return nil, err
	}
	return consumerFromClient, nil
}

func unmarshalIncome(consumerMessage *sarama.ConsumerMessage) (*domain.Message, error) {
	msg := &domain.Message{}
	msg.Data = consumerMessage.Value

	for _, header := range consumerMessage.Headers {
		switch string(header.Key) {
		case "Type":
			msg.Type = string(header.Value)
		case "Message-ID":
			msg.MessageID = domain.UUID{UUID: uuid.FromBytesOrNil(header.Value)}
		case "In-Reply-To":
			msg.InReplyTo = domain.UUID{UUID: uuid.FromBytesOrNil(header.Value)}
		case "Some-about-program":
			msg.Some = string(header.Value)
		}
	}
	return msg, nil // TODO: never return error. need some valdates..
}

// func thatForMe(msg *domain.Message, programID domain.UUID) bool {
// 	if msg.ProgramID == programID {
// 		return true
// 	}
// 	return false
// }

func (saramaPortAdapter *SaramaPortAdapter) addNewJob(jobUUID string) {
	saramaPortAdapter.jobs.mx.Lock()
	saramaPortAdapter.jobs.pool[jobUUID] = struct{}{}
	saramaPortAdapter.jobs.mx.Unlock()
}

func (saramaPortAdapter *SaramaPortAdapter) removeDoneJob(jobUUID string) {
	saramaPortAdapter.jobs.mx.Lock()
	delete(saramaPortAdapter.jobs.pool, jobUUID)
	saramaPortAdapter.jobs.mx.Unlock()
}

// MsgHandler ...
func (saramaPortAdapter *SaramaPortAdapter) MsgHandler() {
	for {
		select {
		case msg := <-saramaPortAdapter.producer.asyncProducer.Successes():
			saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
				"entity":     consumerLogPrefix,
				"event uuid": "00000000-0000-0000-0000-000000000000", // TODO: correct uuid
			}).Tracef("got success send/read message: %v", msg)
		case msg := <-saramaPortAdapter.producer.asyncProducer.Errors():
			saramaPortAdapter.Consumer.logging.WithFields(logrus.Fields{
				"entity":     consumerLogPrefix,
				"event uuid": "00000000-0000-0000-0000-000000000000", // TODO: correct uuid
			}).Tracef("got error send/read message: %v", msg)
		}
	}
}
