package app

import (
	"context"
	"encoding/json"

	"github.com/alejoacosta74/go-logger"
	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string

	outboundChan chan *ChatMessage // messages to be sent to the chat room
	inboundChan  chan *ChatMessage // messages received from the chat room
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message    string
	SenderID   string
	SenderNick string
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*ChatRoom, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ctx:          ctx,
		ps:           ps,
		topic:        topic,
		sub:          sub,
		self:         selfID,
		nick:         nickname,
		roomName:     roomName,
		outboundChan: make(chan *ChatMessage, ChatRoomBufSize),
		inboundChan:  make(chan *ChatMessage, ChatRoomBufSize),
	}

	go cr.eventLoop()
	return cr, nil
}

func (cr *ChatRoom) Publish(message string) error {
	msg := &ChatMessage{
		Message:    message,
		SenderID:   cr.self.String(),
		SenderNick: cr.nick,
	}

	select {
	case cr.outboundChan <- msg:
		return nil
	case <-cr.ctx.Done():
		logger.Warn("context done")
		return cr.ctx.Err()
	}
}

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.roomName))
}

func (cr *ChatRoom) eventLoop() {

	receivedMsgCh := make(chan *pubsub.Message)
	go func() {
		for {
			msg, err := cr.sub.Next(cr.ctx)
			if err != nil {
				logger.Warn("error receiving message", err)
				continue
			}
			if msg.ReceivedFrom == cr.self {
				continue
			}
			receivedMsgCh <- msg
		}
	}()

	for {
		select {
		case <-cr.ctx.Done():
			return
		case msg := <-cr.outboundChan:
			logger.Debug("sending message to chat room")
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				logger.Warn("error marshalling message", err)
				continue
			}
			if err := cr.topic.Publish(cr.ctx, msgBytes); err != nil {
				logger.Warn("error publishing message", err)
			}
		case msg := <-receivedMsgCh:
			if msg.ReceivedFrom == cr.self {
				continue
			}
			cm := new(ChatMessage)
			if err := json.Unmarshal(msg.Data, cm); err != nil {
				logger.Warn("error unmarshalling message", err)
				continue
			}
			select {
			case cr.inboundChan <- cm:
			case <-cr.ctx.Done():
				return
			}
		}
	}
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}
