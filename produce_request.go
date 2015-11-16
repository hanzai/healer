package gokafka

//import (
//"encoding/binary"
//"errors"
//)

/*
The produce API is used to send message sets to the server. For efficiency it allows sending message sets intended for many topic partitions in a single request.
The produce API uses the generic message set format, but since no offset has been assigned to the messages at the time of the send the producer is free to fill in that field in any way it likes.

Produce Request
	ProduceRequest => RequiredAcks Timeout [TopicName [Partition MessageSetSize MessageSet]]
	  RequiredAcks => int16
	  Timeout => int32
	  Partition => int32
	  MessageSetSize => int32

Field			Description
RequiredAcks	This field indicates how many acknowledgements the servers should receive before responding to the request. If it is 0 the server will not send any response (this is the only case where the server will not reply to a request). If it is 1, the server will wait the data is written to the local log before sending a response. If it is -1 the server will block until the message is committed by all in sync replicas before sending a response. For any number > 1 the server will block waiting for this number of acknowledgements to occur (but the server will never wait for more acknowledgements than there are in-sync replicas).
Timeout			This provides a maximum time in milliseconds the server can await the receipt of the number of acknowledgements in RequiredAcks. The timeout is not an exact limit on the request time for a few reasons: (1) it does not include network latency, (2) the timer begins at the beginning of the processing of this request so if many requests are queued due to server overload that wait time will not be included, (3) we will not terminate a local write so if the local write time exceeds this timeout it will not be respected. To get a hard timeout of this type the client should use the socket timeout.
TopicName		The topic that data is being published to.
Partition		The partition that data is being published to.
MessageSetSize	The size, in bytes, of the message set that follows.
MessageSet		A set of messages in the standard format described above.
*/
type ProduceRequest struct {
	RequiredAcks   int16
	Timeout        int32
	Partition      int32
	MessageSetSize int32
}
