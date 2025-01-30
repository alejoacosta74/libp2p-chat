package app

import (
	"context"

	"github.com/alejoacosta74/go-logger"
	uilogger "github.com/alejoacosta74/libp2p-chat-app/logger"
	"github.com/alejoacosta74/libp2p-chat-app/p2p/node"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) error {
	// Create a new libp2p node with the provided context
	p2pNode := node.NewNode(ctx)

	// Initialize the GossipSub pubsub service for p2p message broadcasting
	ps, err := p2pNode.CreatePubSubService()
	if err != nil {
		return err
	}

	// Join the specified chat room using the pubsub service, node ID, and user preferences
	cr, err := JoinChatRoom(ctx, ps, p2pNode.ID(), viper.GetString("nickname"), viper.GetString("room"))
	if err != nil {
		return err
	}

	// Create the terminal UI instance for the chat room
	ui := NewChatUI(cr)
	// Initialize the global UI logger to capture logs in the UI
	uilogger.InitGlobalLogger(ui)
	// Redirect all logger output to the UI logger
	logger.SetOutput(uilogger.GlobalUILogger)
	// If a log file is specified, also write logs to that file
	if viper.GetString("logfile") != "" {
		logger.AddFileOutputHook(viper.GetString("logfile"), nil)
	}

	// Initialize the p2p node, starting discovery services and event listeners
	err = p2pNode.Init()
	if err != nil {
		return err
	}

	// Start the terminal UI event loop and block until exit
	return ui.Run()
}
