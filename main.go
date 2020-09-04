package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jander/golog/logger"
	"github.com/jinzhu/configor"
	"github.com/kardianos/service"
	"net"
	"os"
	"strings"
)

const configPath = "./config.yml"

func init() {
	rotatingHandler := logger.NewRotatingHandler("./", "logsync.log", 1, 1*1024*1024)
	logger.SetHandlers(logger.Console, rotatingHandler)
}

// 配置
type Config struct {
	Topic    string `required:"true" json:"topic"`           // iot/device/{mac地址}/control
	Qos      uint   `required:"true" default:"2" json:"qos"` //0,1,2
	ClientID string `required:"true" json:"clientid"`
	Broker   string `required:"true" default:"tcp://10.0.0.27:1883" json:"broker"`
	UserName string `required:"true" default:"Chindeo" json:"user_name"`
	Password string `required:"true" default:"P@ssw0rd" json:"password"`
	Num      int    `required:"true" default:"1" json:"num"`
}

var config Config
var s service.Service

type program struct {
	client *Client
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	p.client.Start()
	p.udpStart()
}
func (p *program) Stop(s service.Service) error {
	p.client.Stop()
	return nil
}

func main() {
	if err := configor.Load(&config, configPath); err != nil {
		logger.Panicln(fmt.Sprintf("加载配置出错：%v", err))
	}

	svcConfig := &service.Config{
		Name:        "GO_MQTT",
		DisplayName: "设备控制器",
		Description: "MQTT控制设备",
	}

	prg := &program{}
	prg.client = NewClient()
	var err error
	s, err = service.New(prg, svcConfig)
	if err != nil {
		logger.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		logger.Panicln(err)
	}
}

func (p *program) udpStart() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", localIP(), 10006))

	if err != nil {
		logger.Fatal(err)
	}

	go func() {
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Printf("ListenUDP %s:%d", localIP(), 10006)
		for {
			data := make([]byte, 1024)
			_, rAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				logger.Println(fmt.Sprintf("ReadFromUDP %v ", err))
				_, err = conn.WriteToUDP([]byte(err.Error()), rAddr)
				if err != nil {
					logger.Println(err)
					continue
				}
				continue
			}
			err = json.Unmarshal([]byte(strings.Trim(string(data), "\x00")), &config)
			if err != nil {
				logger.Println(fmt.Sprintf(" Unmarshal %v ", err))
				_, err = conn.WriteToUDP([]byte(err.Error()), rAddr)
				if err != nil {
					logger.Println(err)
					continue
				}
				continue
			}

			// 覆盖配置文件
			err = SetConfigFile()
			if err != nil {
				logger.Println(fmt.Sprintf("SetConfigFile %v ", err))
				_, err = conn.WriteToUDP([]byte(err.Error()), rAddr)
				if err != nil {
					logger.Println(err)
					continue
				}
				continue
			}

			p.client.Stop()
			p.client = NewClient()
			if p.client == nil || p.client.MQTT == nil {
				logger.Println("NewClient 失败")
			}
			p.client.Start()

			_, err = conn.WriteToUDP([]byte("配置写入成功"), rAddr)
			if err != nil {
				logger.Println(err)
				continue
			}
		}
	}()
}

func SetConfigFile() error {
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Println("文件打开失败", err)
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)

	configKeys := []string{"topic", "qos", "clientid", "broker", "user_name", "password", "num"}

	for _, key := range configKeys {
		logger.Println("%s : %v", key, config)
		switch key {
		case "topic":
			_, err = write.WriteString(fmt.Sprintf("%s: %s \r\n", key, config.Topic))
			if err != nil {
				return err
			}
		case "qos":
			_, err = write.WriteString(fmt.Sprintf("%s: %d \r\n", key, config.Qos))
			if err != nil {
				return err
			}
		case "clientid":
			_, err = write.WriteString(fmt.Sprintf("%s: %s \r\n", key, config.ClientID))
			if err != nil {
				return err
			}
		case "broker":
			_, err = write.WriteString(fmt.Sprintf("%s: %s \r\n", key, config.Broker))
			if err != nil {
				return err
			}
		case "user_name":
			_, err = write.WriteString(fmt.Sprintf("%s: %s \r\n", key, config.UserName))
			if err != nil {
				return err
			}
		case "password":
			_, err = write.WriteString(fmt.Sprintf("%s: %s \r\n", key, config.Password))
			if err != nil {
				return err
			}
		case "num":
			_, err = write.WriteString(fmt.Sprintf("%s: %d \r\n", key, config.Num))
			if err != nil {
				return err
			}
		}

	}

	err = write.Flush()
	if err != nil {
		return err
	}
	return nil
}
