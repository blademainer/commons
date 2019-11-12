package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/blademainer/commons/pkg/retryer"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"net/url"
	"time"
)

type MqttConnection struct {
	Client mqtt.Client
	Topic  string
}

// Create a MQTT client
func CreateClient(opts *mqtt.ClientOptions) (mqtt.Client, error) {
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return client, token.Error()
	}

	return client, nil
}

// Create a MQTT client
func CreateClientByUri(clientID string, uri *url.URL, keepAlive time.Duration) (mqtt.Client, error) {
	logger.Infof("Create MQTT client and connection: uri=%v clientID=%v", uri.String(), clientID)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
	opts.SetClientID(clientID)
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetConnectTimeout(keepAlive)
	opts.SetKeepAlive(keepAlive)
	opts.SetConnectionLostHandler(reconnectHandler(keepAlive))

	return CreateClient(opts)
}

// Create a MQTT client
func CreateTlsClientByUri(clientID string, uri *url.URL, keepAlive time.Duration, config *tls.Config) (mqtt.Client, error) {
	logger.Infof("Create MQTT client and connection: uri=%v clientID=%v", uri.String(), clientID)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
	opts.SetClientID(clientID)
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetConnectTimeout(keepAlive)
	opts.SetKeepAlive(keepAlive)
	opts.SetConnectionLostHandler(reconnectHandler(keepAlive))
	opts.SetTLSConfig(config)

	return CreateClient(opts)
}

func reconnectHandler(keepaliveTime time.Duration) func(client mqtt.Client, e error) {
	growthRetryer, e := retryer.NewDoubleGrowthRetryer(keepaliveTime)
	if e != nil {
		logger.Errorf("failed to create retryer!")
		return reconnect
	}
	return func(client mqtt.Client, e error) {
		err := growthRetryer.Invoke(func(ctx context.Context) error {
			logger.Warnf("Connection lost : %v", e)
			token := client.Connect()
			if token.Wait() && token.Error() != nil {
				logger.Warnf("Reconnection failed : %v", token.Error())
			} else {
				logger.Warnf("Reconnection sucessful")
			}
			return token.Error()
		})
		if err != nil {
			logger.Errorf("failed to reconnect: %v", err.Error())
		}
	}
}

func reconnect(client mqtt.Client, e error) {
	logger.Warnf("Connection lost : %v", e)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		logger.Warnf("Reconnection failed : %v", token.Error())
	} else {
		logger.Warnf("Reconnection sucessful")
	}
}
