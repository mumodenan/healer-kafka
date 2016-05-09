package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/childe/gokafka"
)

var (
	brokerList = flag.String("brokers", "127.0.0.1:9092", "<hostname:port,...,hostname:port> The comma separated list of brokers in the Kafka cluster. (default: 127.0.0.1:9092)")
	topic      = flag.String("topic", "", "REQUIRED: The topic to get offset from.")
	timeValue  = flag.Int64("time", -1, "timestamp/-1(latest)/-2(earliest). timestamp of the offsets before that.(default: -1) ")
	offsets    = flag.Uint("offsets", 1, "number of offsets returned (default: 1)")
	clientID   = flag.String("clientID", "gokafka", "The ID of this client.")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	flag.Parse()

	if *topic == "" {
		fmt.Println("need topic!")
		flag.PrintDefaults()
	}

	correlationID := os.Getpid()
	metadataResponse, err := gokafka.GetMetaData(*brokerList, *topic, int32(correlationID), *clientID)
	if err != nil {
		logger.Fatal(err)
	}

	brokers := metadataResponse.Brokers

	partitions := metadataResponse.TopicMetadatas[0].PartitionMetadatas

	for _, partition := range partitions {
		partitionID := partition.PartitionId
		leader := partition.Leader

		for _, broker := range brokers {
			if leader == broker.NodeId {
				requestHeader := &gokafka.RequestHeader{
					ApiKey:        gokafka.API_OffsetRequest,
					ApiVersion:    0,
					CorrelationId: int32(correlationID),
					ClientId:      *clientID,
				}

				partitionOffsetRequestInfos := make(map[uint32]*gokafka.PartitionOffsetRequestInfo)
				partitionOffsetRequestInfos[uint32(partitionID)] = &gokafka.PartitionOffsetRequestInfo{
					Time:               *timeValue,
					MaxNumberOfOffsets: uint32(*offsets),
				}
				topicOffsetRequestInfos := make(map[string]map[uint32]*gokafka.PartitionOffsetRequestInfo)
				topicOffsetRequestInfos[*topic] = partitionOffsetRequestInfos

				offsetReqeust := &gokafka.OffsetReqeust{
					RequestHeader: requestHeader,
					ReplicaId:     -1,
					RequestInfo:   topicOffsetRequestInfos,
				}

				payload := offsetReqeust.Encode()

				dialer := net.Dialer{
					Timeout:   time.Second * 5,
					KeepAlive: time.Hour * 2,
				}

				leaderAddr := net.JoinHostPort(broker.Host, strconv.Itoa(int(broker.Port)))
				conn, connErr := dialer.Dial("tcp", leaderAddr)
				if connErr != nil {
					logger.Println(connErr)
					continue
				}
				conn.Write(payload)

				responsePayload := make([]byte, 1024)
				length, _ := conn.Read(responsePayload)

				offsetResponse := &gokafka.OffsetResponse{}
				offsetResponse.Decode(responsePayload[:length])
				for topic, partitions := range offsetResponse.Info {
					for _, partition := range partitions {
						fmt.Printf("%s:%d", topic, partition.Partition)
						for _, offset := range partition.Offset {
							fmt.Printf(":%d", offset)
						}
						fmt.Println()
					}
				}
			}
		}
	}

}