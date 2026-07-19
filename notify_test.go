//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package notify_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/tdrn-org/go-notify"
)

func TestPayloadRegistry(t *testing.T) {
	payloads := &notify.PayloadRegistry[any]{}
	payloads.Add("payload1", &notify.LogPayload{Msg: "payload1#1"})
	payloads.Add("payload2", &notify.LogPayload{Msg: "payload2#1"})
	payloads.Add("payload2", &notify.LogPayload{Msg: "payload2#2"})
	require.Len(t, payloads.Get("payload0"), 0)
	require.Len(t, payloads.Get("payload1"), 1)
	require.Len(t, payloads.Get("payload2"), 2)
}
