package healer

const (
	RESOURCETYPE_UNKNOWN uint8 = iota
	RESOURCETYPE_ANY
	RESOURCETYPE_TOPIC
	RESOURCETYPE_GROUP
	RESOURCETYPE_CLUSTER
	RESOURCETYPE_TRANSACTIONAL_ID
	RESOURCETYPE_DELEGATION_TOKEN
)
