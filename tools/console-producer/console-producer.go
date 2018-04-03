package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/golang/glog"
	"github.com/childe/healer"
)

var (
	config = healer.DefaultProducerConfig()
	topic  = flag.String("topic", "", "REQUIRED: The topic to consume from.")
)

func init() {
	flag.StringVar(&config.BootstrapServers, "brokers", "127.0.0.1:9092", "The list of hostname and port of the server to connect to.")
	flag.StringVar(&config.CompressionType, "compression.type", "none", "defalut:none")
}

func main() {
	flag.Parse()

	if *topic == "" {
		flag.PrintDefaults()
		os.Exit(4)
	}

	producer := healer.NewProducer(*topic, config)

	if producer == nil {
		fmt.Println("could not create producer")
		os.Exit(5)
	}

	var (
		text     []byte = nil
		line     []byte = nil
		isPrefix bool   = true
		err      error  = nil
	)
	reader := bufio.NewReader(os.Stdin)
	for {
		text = nil
		isPrefix = true
		for isPrefix {
			line, isPrefix, err = reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					os.Exit(0)
				}
				glog.Errorf("readline error:%s", err)
				os.Exit(5)
			}
			if text == nil {
				text = line
			} else {
				text = append(text, line...)
			}
		}
		producer.AddMessage(nil, text)
	}
}