package portadapter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"../domain"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

const saramaLogName = "sarama"

// SaramaPortAdapter ..
type SaramaPortAdapter struct {
	uuidgenerator     domain.UUIDgenerator
	programStruct     *domain.ProgramStruct
	producer          *Producer
	Consumer          *Consumer
	HandlerList       map[string]domain.MessageBusHandler
	SendTopic         string
	logging           *logrus.Logger
	consumerJobChanel chan map[string]*sarama.ConsumerMessage
	jobs              *poolOfWorks
	shutdownCommandForListen,
	listenAndHandleDone chan struct{}
}

type poolOfWorks struct {
	mx   sync.Mutex
	rmx  sync.RWMutex // read mutex
	pool map[string]struct{}
}

// Subscribe recived message
func (saramaPortAdapter *SaramaPortAdapter) Subscribe(msgType string, handler domain.MessageBusHandler) {
	saramaPortAdapter.HandlerList[msgType] = handler
}

// NewSaramaPortAdapter ...
func NewSaramaPortAdapter(
	uuidgenerator domain.UUIDgenerator,
	programStruct *domain.ProgramStruct,
	bootstrapServers,
	topicListen []string, groupID,
	sendTopic string,
	uuidForNewSaramaPortAdapter string,
	monitoring domain.PrometheusConfig,
	jobLimit int,
	logging *logrus.Logger) (*SaramaPortAdapter, error) {

	ticker := time.NewTicker(time.Duration(500 * time.Millisecond))
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()

	createNewSaramaPortAdapter := make(chan struct{}, 1)
	createNewSaramaPortAdapter <- struct{}{}

	for {
		select {
		case <-createNewSaramaPortAdapter:
			producer, err := newProducer(bootstrapServers, uuidForNewSaramaPortAdapter, logging)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     saramaLogName,
					"event uuid": uuidForNewSaramaPortAdapter,
				}).Warnf("can't create producer: %v", err)
				continue
			}

			consumer, err := newConsumer(bootstrapServers, topicListen, groupID, uuidForNewSaramaPortAdapter, monitoring, logging)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     saramaLogName,
					"event uuid": uuidForNewSaramaPortAdapter,
				}).Warnf("can't create consumer: %v", err)
				continue
			}

			jobs := &poolOfWorks{pool: make(map[string]struct{})}

			shutdownCommandForListen := make(chan struct{}, 1)
			listenAndHandleDone := make(chan struct{}, 1)

			return &SaramaPortAdapter{
				uuidgenerator: uuidgenerator,
				programStruct: programStruct,
				producer:      producer,
				Consumer:      consumer,
				HandlerList:   make(map[string]domain.MessageBusHandler),
				SendTopic:     sendTopic,
				logging:       logging,

				consumerJobChanel:        make(chan map[string]*sarama.ConsumerMessage, jobLimit),
				jobs:                     jobs,
				shutdownCommandForListen: shutdownCommandForListen,
				listenAndHandleDone:      listenAndHandleDone,
			}, nil
		case tooLong := <-ctx.Done():
			return nil, fmt.Errorf("can't create producer and consumer in time: %v", tooLong)
		case <-ticker.C:
			createNewSaramaPortAdapter <- struct{}{}
		}
	}
}

func uuidToBytes(uuid domain.UUID) []byte {
	return uuid.UUID.Bytes()
}

func timeToUnixNanoAndToBytes(t time.Time) []byte {
	return []byte(strconv.FormatInt(t.UnixNano()/1e6, 10))
}
