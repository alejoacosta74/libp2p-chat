package app

import (
	"context"
	"encoding/json"

	"github.com/alejoacosta74/libp2p-chat-app/logger"
	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	// Messages is a channel of messages received from other peers in the chat room
	// Messages chan *ChatMessage

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
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		// Messages: make(chan *ChatMessage, ChatRoomBufSize),
		outboundChan: make(chan *ChatMessage, ChatRoomBufSize), // Initialize channel
		inboundChan:  make(chan *ChatMessage, ChatRoomBufSize), // Initialize channel
	}

	// start reading messages from the subscription in a loop
	// go cr.readLoop()
	go cr.eventLoop()
	return cr, nil
}

// Publish sends a message to the pubsub topic.
// func (cr *ChatRoom) Publish(message string) error {
// 	m := ChatMessage{
// 		Message:    message,
// 		SenderID:   cr.self.String(),
// 		SenderNick: cr.nick,
// 	}
// 	msgBytes, err := json.Marshal(m)
// 	if err != nil {
// 		return err
// 	}
// 	return cr.topic.Publish(cr.ctx, msgBytes)
// }

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
		logger.GlobalUILogger.Log("context done")
		return cr.ctx.Err()
	}
}

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.roomName))
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
// func (cr *ChatRoom) readLoop() {
// 	for {
// 		msg, err := cr.sub.Next(cr.ctx)
// 		if err != nil {
// 			close(cr.Messages)
// 			return
// 		}
// 		// only forward messages delivered by others
// 		if msg.ReceivedFrom == cr.self {
// 			continue
// 		}
// 		cm := new(ChatMessage)
// 		err = json.Unmarshal(msg.Data, cm)
// 		if err != nil {
// 			continue
// 		}
// 		// send valid messages onto the Messages channel
// 		cr.Messages <- cm
// 	}
// }

func (cr *ChatRoom) eventLoop() {

	receivedMsgCh := make(chan *pubsub.Message)
	go func() {
		for {
			msg, err := cr.sub.Next(cr.ctx)
			if err != nil {
				logger.GlobalUILogger.Log("error receiving message", err)
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
			logger.GlobalUILogger.Log("sending message to chat room")
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				logger.GlobalUILogger.Log("error marshalling message", err)
				continue
			}
			if err := cr.topic.Publish(cr.ctx, msgBytes); err != nil {
				logger.GlobalUILogger.Log("error publishing message", err)
			}
		case msg := <-receivedMsgCh:
			if msg.ReceivedFrom == cr.self {
				continue
			}
			cm := new(ChatMessage)
			if err := json.Unmarshal(msg.Data, cm); err != nil {
				logger.GlobalUILogger.Log("error unmarshalling message", err)
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
