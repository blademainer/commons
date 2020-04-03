package mqtt

import (
	"errors"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/blademainer/commons/pkg/rpc/queue"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

type mqttQueue struct {
	mqtt.Client
	sync.Once

	qos           byte
	retain        bool
	consumeTopics []string
	produceTopic  string
	payloadCh     chan []byte
}

func (m *mqttQueue) Produce(payload []byte) error {
	token := m.Publish(m.produceTopic, m.qos, m.retain, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *mqttQueue) Consume() (payload []byte, e error) {
	// check go routine
	m.Do(m.startSubscribe)
	select {
	case payload, open := <-m.payloadCh:
		if !open {
			e = errors.New("chan is closed")
			return payload, e
		}
		return payload, e
	}
}

func (m *mqttQueue) startSubscribe() {
	if len(m.consumeTopics) == 0 {
		logger.Fatalf("consume topics is empty")
	}
	for _, topic := range m.consumeTopics {
		logger.Infof("sub topic: %v qos: %v", topic, m.qos)
		m.Client.Subscribe(topic, m.qos, func(client mqtt.Client, message mqtt.Message) {
			if logger.IsDebugEnabled() {
				logger.Debugf("receive message from topic: %v, message: %v", topic, message)
			}
			defer message.Ack()
			payload := message.Payload()
			m.payloadCh <- payload
		})
	}
}

func (m *mqttQueue) Close() (e error) {
	close(m.payloadCh)
	m.Client.Disconnect(5000)
	return
}

func NewMqttQueue(client mqtt.Client, qos byte, retain bool, produceTopic string, consumeTopics ...string) queue.Queue {
	q := &mqttQueue{}
	q.Client = client
	q.qos = qos
	q.retain = retain
	q.produceTopic = produceTopic
	q.consumeTopics = consumeTopics
	q.payloadCh = make(chan []byte, 1024)
	return q
}
