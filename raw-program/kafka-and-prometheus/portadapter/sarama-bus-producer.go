package portadapter

import (
	"time"

	"../domain"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const producerLogPrefix = "sarama.producer"

var producerConfig *sarama.Config

func init() {
	producerConfig = sarama.NewConfig()
	producerConfig.Version = sarama.V2_1_0_0
	producerConfig.Producer.Return.Errors = false    //  passive parameter, no need any more
	producerConfig.Producer.Return.Successes = false //  passive parameter, no need any more
}

// Producer ...
type Producer struct {
	asyncProducer sarama.AsyncProducer
	logging       *logrus.Logger
}

func newProducer(bootstrapServers []string, uuidForNewSaramaPortAdapter string, logging *logrus.Logger) (*Producer, error) {
	client, err := newProducerClient(bootstrapServers)
	if err != nil {
		return nil, err
	}

	asyncProducer, err := newAsyncProducer(client) // that will log not here
	if err != nil {
		return nil, err
	}

	logging.WithFields(logrus.Fields{
		"entity":     producerLogPrefix,
		"event uuid": uuidForNewSaramaPortAdapter,
	}).Info("successfully create producer client")

	producer := &Producer{
		asyncProducer: asyncProducer,
	}
	return producer, nil
}

// Send message to kafka
func (saramaPortAdapter *SaramaPortAdapter) Send(message domain.Message) {
	pmsg := saramaPortAdapter.prepareMessageForSend(message)
	saramaPortAdapter.producer.asyncProducer.Input() <- pmsg
	saramaPortAdapter.logging.WithFields(logrus.Fields{
		"entity":     producerLogPrefix,
		"event uuid": message.InternalUUID,
	}).Debug("successfully send message to producer chan")
}

func newProducerClient(bootstrapServers []string) (sarama.Client, error) {
	producerConfig.Net.WriteTimeout = time.Duration(30 * time.Second)
	producerClient, err := sarama.NewClient(bootstrapServers, producerConfig)
	if err != nil {
		return nil, err
	}
	return producerClient, nil
}

func newAsyncProducer(producerClient sarama.Client) (sarama.AsyncProducer, error) {
	asyncProducer, err := sarama.NewAsyncProducerFromClient(producerClient)
	if err != nil {
		return nil, err
	}
	return asyncProducer, nil
}

func (saramaPortAdapter *SaramaPortAdapter) prepareMessageForSend(msg domain.Message) *sarama.ProducerMessage {
	//endpointID domain.UUID, data interface{}, msgType, contentType string) *sarama.ProducerMessage {
	pmsg := &sarama.ProducerMessage{Topic: saramaPortAdapter.SendTopic}

	headers := map[string][]byte{
		"Message-ID":         uuidToBytes(saramaPortAdapter.uuidgenerator.NewUUID()),
		"To":                 []byte(saramaPortAdapter.SendTopic),
		"In-Reply-To":        uuidToBytes(msg.InReplyTo),
		"Type":               []byte("sometype"),
		"Some-about-program": []byte(saramaPortAdapter.programStruct.SomeData),
	}

	for k, v := range headers {
		pmsg.Headers = append(pmsg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: v,
		})
	}

	dataBytes, isConverted := msg.Data.([]byte)
	if !isConverted { // if don't use that - if dataBytes not bytes -> panic // TODO: FIXME: wgat?
		// log.Error("data is not bytes, some go wrong")
	}
	pmsg.Value = sarama.StringEncoder(dataBytes)
	saramaPortAdapter.logging.WithFields(logrus.Fields{
		"entity":     producerLogPrefix,
		"event uuid": msg.InternalUUID,
	}).Debugf("successfully prepared message %v for send", msg.InternalUUID)

	return pmsg
}
