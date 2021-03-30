package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mumodenan/healer-kafka"
	"github.com/golang/glog"
)

var (
	brokerConfig = healer.DefaultBrokerConfig()

	brokerList = flag.String("brokers", "127.0.0.1:9092", "REQUIRED: The list of hostname and port of the server to connect to.")
	clientID   = flag.String("clientID", "healer", "The ID of this client.")
)

func init() {
	flag.IntVar(&brokerConfig.ConnectTimeoutMS, "connect-timeout", brokerConfig.ConnectTimeoutMS, fmt.Sprintf("connect timeout to broker. default %d", brokerConfig.ConnectTimeoutMS))
	flag.IntVar(&brokerConfig.TimeoutMS, "timeout", brokerConfig.TimeoutMS, fmt.Sprintf("read timeout from connection to broker. default %d", brokerConfig.TimeoutMS))
}

func main() {
	flag.Parse()

	helper, err := healer.NewHelper(*brokerList, *clientID, brokerConfig)
	if err != nil {
		glog.Error("create helper error:%s", err)
		os.Exit(5)
	}

	groups := helper.GetGroups()
	if groups != nil {
		for _, group := range groups {
			fmt.Println(group)
		}
	}
}
