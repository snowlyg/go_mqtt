package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func localIP() string {
	ip := ""
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsMulticast() && !ipnet.IP.IsLinkLocalUnicast() && !ipnet.IP.IsLinkLocalMulticast() && ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}
		}
	}
	return ip
}

type Config struct {
	Topic    string `required:"true" json:"topic"`           // iot/device/{mac地址}/control
	Qos      uint   `required:"true" default:"2" json:"qos"` //0,1,2
	ClientID string `required:"true" json:"clientid"`
	Broker   string `required:"true" default:"tcp://10.0.0.27:1883" json:"broker"`
	UserName string `required:"true" default:"Chindeo" json:"user_name"`
	Password string `required:"true" default:"P@ssw0rd" json:"password"`
	Num      int    `required:"true" default:"1" json:"num"`
}

func main() {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", localIP(), 10006))
	checkError(err)

	defer conn.Close()

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		_, err := conn.Write([]byte(line))
		checkError(err)

		fmt.Println("Write:", string(line))

		msg := make([]byte, 1024)
		_, err = conn.Read(msg)
		checkError(err)

		fmt.Println("Response:", string(msg))
	}
}
