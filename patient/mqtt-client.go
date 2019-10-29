package patient

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient struct {
	client mqtt.Client
}

var E_MqttClient *MqttClient

func InitMqttClient() (err error) {
	var (
		opts   *mqtt.ClientOptions
		client mqtt.Client
		token  mqtt.Token
	)

	opts = mqtt.NewClientOptions().
		AddBroker(E_Config.BrokerIp).                // broker ip
		SetClientID(E_Config.MqttClientId).          // client id
		SetProtocolVersion(E_Config.MqttProtocolVer) // mqtt 3.1.1
		//SetWill(E_Config.MqttClientId + "/will", "Goodbye", 0, true).	// LWT临终遗嘱连接，客户端离线后会有will消息
		//SetUsername(E_Config.MqttUserName).													// 用户名和密码
		//SetPassword(E_Config.MqttUserPassword)

	client = mqtt.NewClient(opts)

	// 开启连接
	if token = client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	E_MqttClient = &MqttClient{
		client: client,
	}

	return nil
}

// Pub 发布消息到指定主题
func (c *MqttClient) Pub(topic string, payload string) (err error) {
	if token := c.client.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Sub 订阅消息主题
func (c *MqttClient) Sub(topic string) (err error) {
	//msgRcvd := func(client *mqtt.Client, message mqtt.Message) {
	//	fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	//}
	//
	//if token := c.client.Subscribe("example/topic", 0, msgRcvd); token.Wait() && token.Error() != nil {
	//	fmt.Println(token.Error())
	//}

	if token := c.client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// 取消订阅主题
func (c *MqttClient) UnSub(topic string) (err error) {
	if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// 取消连接
func (c *MqttClient) DisConnect(quiesce uint) {
	c.client.Disconnect(quiesce)
}

// 建立连接
// TODO
func (c *MqttClient) Connect() mqtt.Token {
	return nil
}
