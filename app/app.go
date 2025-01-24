package app

import (
	"context"

	"github.com/spf13/viper"
)

func Run(ctx context.Context) error {
	node := NewNode(ctx)
	err := node.SetupDiscovery()
	if err != nil {
		return err
	}

	ps, err := node.CreatePubSubService()
	if err != nil {
		return err
	}

	cr, err := JoinChatRoom(ctx, ps, node.ID(), viper.GetString("nickname"), viper.GetString("room"))
	if err != nil {
		return err
	}

	ui := NewChatUI(cr)
	return ui.Run()
}
