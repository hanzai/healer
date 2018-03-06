package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/childe/healer"
	"github.com/golang/glog"
)

var (
	bootstrapServers = flag.String("bootstrap.servers", "127.0.0.1:9092", "The list of hostname and port of the server to connect to(defautl: 127.0.0.1:9092).")
	topic            = flag.String("topic", "", "get all topics if not given")
	clientID         = flag.String("clientID", "", "The ID of this client.")
	groupID          = flag.String("groupID", "", "")

	connectTimeout = flag.Int("connect-timeout", 30, "default 30 Second. connect timeout to broker")
	timeout        = flag.Int("timeout", 60, "default 60 Second. read timeout from connection to broker")
)

var (
	brokers *healer.Brokers
	err     error
)

func getPartitions(topic string) ([]int32, error) {
	var metadataResponse *healer.MetadataResponse
	metadataResponse, err = brokers.RequestMetaData(*clientID, []string{topic})

	if err != nil {
		return nil, err
	}

	partitions := make([]int32, 0)
	for _, topicMetadata := range metadataResponse.TopicMetadatas {
		for _, partitionMetadata := range topicMetadata.PartitionMetadatas {
			partitions = append(partitions, partitionMetadata.PartitionID)
		}
	}

	return partitions, nil
}

func getOffset(topic string) (map[int32]int64, error) {
	var (
		partitionID int32 = -1
		timestamp   int64 = -1
	)
	offsetsResponses, err := brokers.RequestOffsets(*clientID, topic, partitionID, timestamp, 1)
	if err != nil {
		return nil, err
	}

	rst := make(map[int32]int64)
	for _, offsetsResponse := range offsetsResponses {
		for _, partitionOffsets := range offsetsResponse.TopicPartitionOffsets {
			for topic, partitionOffset := range partitionOffsets {
				if len(partitionOffset.Offsets) != 1 {
					return nil, fmt.Errorf("%s[%d] offsets return more than 1 value", topic, partitionOffset.Partition)
				}
				rst[partitionOffset.Partition] = partitionOffset.Offsets[0]
			}
		}
	}
	return rst, nil
}

func getCommitedOffset(topic string, partitions []int32) (map[int32]int64, error) {
	coordinatorResponse, err := brokers.FindCoordinator(*clientID, *groupID)
	if err != nil {
		return nil, err
	}

	coordinator, err := brokers.GetBroker(coordinatorResponse.Coordinator.NodeID)
	if err != nil {
		return nil, err
	}
	glog.Infof("coordinator:%s", coordinator.GetAddress())

	r := healer.NewOffsetFetchRequest(1, *clientID, *groupID)
	for _, p := range partitions {
		r.AddPartiton(topic, p)
	}

	response, err := coordinator.Request(r)
	if err != nil {
		return nil, err
	}

	res, err := healer.NewOffsetFetchResponse(response)
	if err != nil {
		return nil, err
	}

	rst := make(map[int32]int64)
	for _, t := range res.Topics {
		for _, p := range t.Partitions {
			rst[p.PartitionID] = p.Offset
		}
	}
	return rst, nil
}

func getAllTopics() ([]string, error) {
	var metadataResponse *healer.MetadataResponse
	metadataResponse, err = brokers.RequestMetaData(*clientID, nil)

	if err != nil {
		return nil, err
	}

	topics := make([]string, 0)
	for _, t := range metadataResponse.TopicMetadatas {
		topics = append(topics, t.TopicName)
	}

	return topics, nil
}

func main() {
	flag.Parse()

	brokers, err = healer.NewBrokers(*bootstrapServers, *clientID, *connectTimeout, *timeout)
	if err != nil {
		glog.Errorf("create brokers error:%s", err)
		os.Exit(5)
	}

	var topics []string
	if *topic == "" {
		topics, err = getAllTopics()
		if err != nil {
			glog.Errorf("fetch topics error:%s", err)
			os.Exit(5)
		}
	} else {
		topics = []string{*topic}
	}

	if *groupID == "" {
		flag.PrintDefaults()
		fmt.Println("need group name")
		os.Exit(4)
	}

	for _, topicName := range topics {
		partitions, err := getPartitions(topicName)
		if err != nil {
			glog.Errorf("get partitions error:%s", err)
			os.Exit(5)
		}

		offsets, err := getOffset(topicName)
		if err != nil {
			glog.Errorf("get offsets error:%s", err)
			os.Exit(5)
		}

		commitedOffsets, err := getCommitedOffset(topicName, partitions)
		if err != nil {
			glog.Errorf("get commitedOffsets error:%s", err)
			os.Exit(5)
		}

		fmt.Println("topic\tpid\toffset\tcommited\tlag")
		var (
			offsetSum   int64 = 0
			commitedSum int64 = 0
			pendingSum  int64 = 0
		)

		for _, partitionID := range partitions {
			pending := offsets[partitionID] - commitedOffsets[partitionID]
			offsetSum += offsets[partitionID]
			commitedSum += commitedOffsets[partitionID]
			pendingSum += pending
			fmt.Printf("%s\t%d\t%d\t%d\t%d\n", topicName, partitionID, offsets[partitionID], commitedOffsets[partitionID], pending)
		}
		fmt.Printf("total:\t%d\t%d\t%d\n", offsetSum, commitedSum, pendingSum)
	}
}
