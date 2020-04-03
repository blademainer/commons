package mqtt

import "fmt"

// Device read message topic
var inboxFormat = "v1/node/%s/inbox"

// Device read message topic
//var inboxFormat = "v1/device/%v/inbox"

// Device publish message topic
var outgoingFormat = "v1/node/%s/outbox"

// Topic of subscribe all nodes
var allNodesOutgoingFormat = "v1/node/+/outbox"

// Server routing to devices topic
var serverPublishTopic = "v1/global/publish"

// Devices sending to defaultServer
var serverSubscribeTopic = "v1/global/subscribe"

type MqttRouterConfig struct {
	Path   string
	Server interface{}
}

type MqttCmd struct {
	Service string
	Method  string
}

func ControllerListenTopic(groupId string) string {
	return fmt.Sprintf("$share/%v/%v", groupId, serverSubscribeTopic)
}

func ConsumeAllNodesTopic(groupId string) string {
	return fmt.Sprintf("$share/%v/%v", groupId, allNodesOutgoingFormat)
}

// The topic which is the message broadcasting to nodes.
func ServerPublishTopic() string {
	return serverPublishTopic
}

// The topic which is the message sent from node
func NodeOutgoingTopic(nodeId string) string {
	return fmt.Sprintf(outgoingFormat, nodeId)
}

// The topic which is the message sent to node
func NodeInboxTopic(nodeId string) string {
	return fmt.Sprintf(inboxFormat, nodeId)
}
