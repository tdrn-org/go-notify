//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mattermost

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
)

type Config interface {
	ServerURL() string
	BotToken() string
	ChannelID() string
}

type PayloadFactory struct {
	client    *model.Client4
	channelID string
}

func NewPayloadFactory(config Config) *PayloadFactory {
	client := model.NewAPIv4Client(config.ServerURL())
	client.SetToken(config.BotToken())
	payloadFactory := &PayloadFactory{
		client:    client,
		channelID: config.ChannelID(),
	}
	return payloadFactory
}

func (f *PayloadFactory) CreatePayload(post *Post) *Payload {
	payload := &Payload{
		factory: f,
		post:    post,
	}
	return payload
}

type Post struct {
	Message string
}

type Payload struct {
	factory *PayloadFactory
	post    *Post
}

func (p *Payload) Send(ctx context.Context, params any) error {
	post := &model.Post{
		ChannelId: p.factory.channelID,
		Message:   p.post.Message,
	}
	_, _, err := p.factory.client.CreatePost(ctx, post)
	if err != nil {
		return fmt.Errorf("failed to send mattermost notification (cause: %w)", err)
	}
	return nil
}
