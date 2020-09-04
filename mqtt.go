package main

import (
	"encoding/json"
	"fmt"
	"github.com/jander/golog/logger"
	"net"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-cmd/cmd"
)

type Client struct {
	MQTT   MQTT.Client
	Opts   *MQTT.ClientOptions
	Choke  chan [2]string
	Borker string
}

func NewClient() *Client {
	var client Client
	client.NewMqttCleintOptions()
	newClient := MQTT.NewClient(client.Opts)
	client.MQTT = newClient
	return &client
}

func (c *Client) NewMqttCleintOptions() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetClientID(config.ClientID)
	opts.SetUsername(config.UserName)
	opts.SetPassword(config.Password)
	opts.SetCleanSession(true)
	choke := make(chan [2]string)
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	c.Opts = opts
	c.Choke = choke

}

// 消息
type Msg struct {
	Type string `json:"type"` //cmd
	Data struct {
		Text string `json:"text"` // 执行的shell脚本
		Arg  string `json:"arg"`  // 执行的shell脚本参数
	} `json:"data"`
}

func (c *Client) Start() {
	go func() {
		logger.Printf("MQTTClient Start with config: %v", config)
		receiveCount := 0
		client := c.MQTT
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			logger.Panicln(fmt.Sprintf("Connect %v", token.Error()))
		}

		logger.Printf("MQTTClient Connect with config: %v", config)

		topic := fmt.Sprintf("iot/device/%s/control", macAddr())
		logger.Printf("MQTTClient Connect topic: %s", topic)

		if token := client.Subscribe(topic, byte(config.Qos), nil); token.Wait() && token.Error() != nil {
			logger.Panicln(fmt.Sprintf("Subscribe %v", token.Error()))
		}

		logger.Printf("MQTTClient Subscribe with config: %v", config)

		for receiveCount < config.Num {
			incoming := <-c.Choke
			topic := incoming[0]
			msg := incoming[1]
			var m Msg
			if err := json.Unmarshal([]byte(msg), &m); err != nil {
				logger.Printf("消息反JSON出错: %v\n", err)
			}

			switch m.Type {
			case "cmd": // 执行 cmd
				logger.Println("cmd start")
				args := strings.Split(m.Data.Arg, " ")
				cmdOptions := cmd.Options{
					Buffered:  true,
					Streaming: true,
				}

				envCmd := cmd.NewCmdOptions(cmdOptions, m.Data.Text, args...)
				go func() {
					for envCmd.Stdout != nil || envCmd.Stderr != nil {
						select {
						case line, open := <-envCmd.Stdout:
							if !open {
								envCmd.Stdout = nil
								continue
							}
							logger.Println(line)
						case line, open := <-envCmd.Stderr:
							if !open {
								envCmd.Stderr = nil
								continue
							}
							logger.Println(os.Stderr, line)
						}
					}
				}()

				envCmd.Start()
			}

			receiveCount++

			logger.Printf("主题：%s\n", topic)
			logger.Printf("消息内容：%s\n", msg)
		}
	}()

}

func (c *Client) Stop() {
	c.MQTT.Disconnect(250)
	logger.Println("MQTT Subscriber Disconnected")
}

func (c *Client) SetOpts() {
	c.Stop()
	c.Opts.AddBroker(config.Broker)
	c.Opts.SetClientID(config.ClientID)
	c.Opts.SetUsername(config.UserName)
	c.Opts.SetPassword(config.Password)
	c.Opts.SetCleanSession(false)
	choke := make(chan [2]string)

	c.Opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	c.Choke = choke
	c.Start()
}

func macAddr() string {
	var macAddrs []string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs[1]
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return strings.Replace(macAddrs[1], ":", "", 5)
}
