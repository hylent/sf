package clients

import (
	"fmt"
	"github.com/aliyunmq/mq-http-go-sdk"
	"strings"
)

type RocketMqMessage struct {
	Id            string
	Body          string
	Tag           string
	PublishMilli  int64
	ConsumedTimes int64
}

type RocketMqConsumer func([]*RocketMqMessage) error

type RocketMqPub struct {
	Topic string `yaml:"topic"`

	handler mq_http_sdk.MQProducer
}

type RocketMqSub struct {
	Topic           string `yaml:"topic"`
	GroupId         string `yaml:"group_id"`
	Tag             string `yaml:"tag"`
	PeekNum         int32  `yaml:"peek_num"`
	PeekWaitSeconds int64  `yaml:"peek_wait_seconds"`
	Orderly         bool   `yaml:"orderly"`

	handler mq_http_sdk.MQConsumer
}

type RocketMqClient struct {
	Endpoint        string                 `yaml:"endpoint"`
	AccessKeyId     string                 `yaml:"access_key_id"`
	AccessKeySecret string                 `yaml:"access_key_secret"`
	InstanceId      string                 `yaml:"instance_id"`
	Pubs            map[string]RocketMqPub `yaml:"pubs"`
	Subs            map[string]RocketMqSub `yaml:"subs"`

	client mq_http_sdk.MQClient
}

func (x *RocketMqClient) Init() error {
	x.client = mq_http_sdk.NewAliyunMQClient(
		x.Endpoint,
		x.AccessKeyId,
		x.AccessKeySecret,
		"",
	)

	for _, item := range x.Pubs {
		item.handler = x.client.GetProducer(x.InstanceId, item.Topic)
	}
	for _, item := range x.Subs {
		item.handler = x.client.GetConsumer(x.InstanceId, item.Topic, item.GroupId, item.Tag)
	}

	return nil
}

func (x *RocketMqClient) Pub(name string, body string, tag string, shardingKey string) (string, error) {
	pub, pubFound := x.Pubs[name]
	if !pubFound {
		return "", fmt.Errorf("rmq_invalid_pub: name=%s", name)
	}

	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: body,
		MessageTag:  tag,
		ShardingKey: shardingKey,
	}
	ret, err := pub.handler.PublishMessage(msg)
	if err != nil {
		return "", fmt.Errorf("rmq_msg_pub_fail: err=%v", err)
	}

	return ret.MessageId, nil
}

func (x *RocketMqClient) Sub(name string, consumer RocketMqConsumer) error {
	sub, subFound := x.Subs[name]
	if !subFound {
		return fmt.Errorf("rmq_invalid_sub: name=%s", name)
	}

	var ackIdList []string
	var msgList []*RocketMqMessage

	respChan := make(chan mq_http_sdk.ConsumeMessageResponse)
	errChan := make(chan error)

	if sub.Orderly {
		go sub.handler.ConsumeMessageOrderly(
			respChan,
			errChan,
			sub.PeekNum,
			sub.PeekWaitSeconds,
		)
	} else {
		go sub.handler.ConsumeMessage(
			respChan,
			errChan,
			sub.PeekNum,
			sub.PeekWaitSeconds,
		)
	}

	select {
	case err := <-errChan:
		if strings.Contains(err.Error(), "MessageNotExist") {
			return nil
		}
		return fmt.Errorf("rmq_peek_fail: err=%v", err)

	case resp := <-respChan:
		for _, v := range resp.Messages {
			ackIdList = append(ackIdList, v.ReceiptHandle)
			msgList = append(msgList, &RocketMqMessage{
				Id:            v.MessageId,
				Body:          v.MessageBody,
				Tag:           v.MessageTag,
				PublishMilli:  v.PublishTime,
				ConsumedTimes: v.ConsumedTimes,
			})
		}
	}

	if err := consumer(msgList); err != nil {
		return fmt.Errorf("rmq_consume_fail: err=%v", err)
	}

	if len(ackIdList) > 0 {
		if err := sub.handler.AckMessage(ackIdList); err != nil {
			return fmt.Errorf("rmq_ack_fail: ackIdList=%+v err=%v", ackIdList, err)
		}
	}

	return nil
}
