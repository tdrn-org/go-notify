//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mattermost_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/tdrn-org/go-notify/mattermost"
)

func TestMattermost(t *testing.T) {
	config := newMattermostConfig()
	if config == nil {
		t.SkipNow()
	}
	factory, err := mattermost.NewPayloadFactory(config)
	require.NoError(t, err)
	message := "Hello world from {{.Testname}}!"
	payload := factory.NewPayload(message)
	params := &Params{
		Testname: t.Name(),
	}
	err = payload.Send(t.Context(), params)
	require.NoError(t, err)
}

type Params struct {
	Testname string
}

func newMattermostConfig() mattermost.Config[*Params] {
	return nil
}
