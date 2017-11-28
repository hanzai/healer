package healer

import "sort"

type AssignmentStrategy interface {
	// generally topicMetadatas is returned by metaDataRequest sent by GroupConsumer
	Assign([]*Member, []*TopicMetadata) GroupAssignment
}

type RangeAssignmentStrategy struct {
}

/*
type PartitionMetadataInfo struct {
	PartitionErrorCode int16
	PartitionId        uint32
	Leader             int32
	Replicas           []int32
	Isr                []int32
}

type TopicMetadata struct {
	TopicErrorCode     int16
	TopicName          string
	PartitionMetadatas []*PartitionMetadataInfo
}

type PartitionAssignment struct {
	Topic     string
	Partition int32
}
type MemberAssignment struct {
	Version              int16
	PartitionAssignments []*PartitionAssignment
	UserData             []byte
}

type GroupAssignment []struct {
	MemberID         string
	MemberAssignment []byte
}
*/

// partitions in one topic
// (3,5)=>[2,2,1]  (5,10)=>[2,2,2,2,2]
func (r *RangeAssignmentStrategy) assignPartitions(memberCount, partitionCount int) []int {
	rst := make([]int, memberCount)
	for i := memberCount - 1; i >= 0; i-- {
		rst[i] = partitionCount / (i + 1)
		partitionCount -= rst[i]
	}

	return rst
}

type ByPartitionID []*PartitionMetadataInfo

func (a ByPartitionID) Len() int           { return len(a) }
func (a ByPartitionID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPartitionID) Less(i, j int) bool { return a[i].PartitionId < a[j].PartitionId }

func (r *RangeAssignmentStrategy) Assign(members []*Member, topicMetadatas []*TopicMetadata) GroupAssignment {

	groupAssignment := GroupAssignment{}
	//for _, member := range members {
	//memberAssignment := &MemberAssignment{
	//Version:  0,
	//UserData: nil,
	//}
	//}

	for _, topicMetadata := range topicMetadatas {
		sort.Sort(ByPartitionID(topicMetadata.PartitionMetadatas))
		partitions := r.assignPartitions(len(members), len(topicMetadata.PartitionMetadatas))
		partitionAssignments := make([]*PartitionAssignment, len(topicMetadata.PartitionMetadatas))
		idx := 0
		for i := 0; i < len(members); i++ {
			for j := 0; j < partitions[i]; j++ {
				partitionAssignments[idx] = &PartitionAssignment{
					Topic:     topicMetadata.TopicName,
					Partition: int32(topicMetadata.PartitionMetadatas[idx].PartitionId),
				}
				idx++
			}
		}
	}

	//groupAssignments[i] = &GroupAssignment{
	//MemberID:         member,
	//MemberAssignment: memberAssignment.Encode(),
	//}
	//}

	return groupAssignment
}
