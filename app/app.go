package app

import (
	"context"

	"github.com/alejoacosta74/go-logger"
	uilogger "github.com/alejoacosta74/libp2p-chat-app/logger"
	"github.com/alejoacosta74/libp2p-chat-app/p2p/node"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) error {
	p2pNode := node.NewNode(ctx)

	ps, err := p2pNode.CreatePubSubService()
	if err != nil {
		return err
	}

	cr, err := JoinChatRoom(ctx, ps, p2pNode.ID(), viper.GetString("nickname"), viper.GetString("room"))
	if err != nil {
		return err
	}

	ui := NewChatUI(cr)
	uilogger.InitGlobalLogger(ui)
	logger.SetOutput(uilogger.GlobalUILogger)

	err = p2pNode.Init()
	if err != nil {
		return err
	}

	return ui.Run()
}
