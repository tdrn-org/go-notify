//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package notify

import (
	"context"
	"errors"
	"sync"
)

// Payload represents a single notification including content and transport
// specific attributes.
type Payload[T any] interface {
	// TransportName gets the name of the underlying transport.
	TransportName() string
	// Send sends the Payload after applying the given params object.
	Send(ctx context.Context, params any) error
}

// Payloads represents an array of [Payload] instances.
type Payloads[T any] []Payload[T]

// Send invokes [Payload.Send] for all given [Payload] instances.
func (payloads Payloads[T]) Send(ctx context.Context, params any) error {
	sendErrs := make([]error, 0, len(payloads))
	for _, payload := range payloads {
		sendErr := payload.Send(ctx, params)
		if sendErr != nil {
			sendErrs = append(sendErrs, sendErr)
		}
	}
	return errors.Join(sendErrs...)
}

// PayloadRegistry represents a registry of named [Payloads] and by this
// a single location to store and access the notifications of an application.
type PayloadRegistry[T any] struct {
	payloads map[string][]Payload[T]
	mutex    sync.RWMutex
}

// Add adds a the given [Payload] instance using the given name. Multiple
// payloads can be defined for a single name.
func (r *PayloadRegistry[T]) Add(name string, payload Payload[T]) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.payloads == nil {
		r.payloads = make(map[string][]Payload[T])
	}
	namedPayloads := r.payloads[name]
	namedPayloads = append(namedPayloads, payload)
	r.payloads[name] = namedPayloads
}

// Get gets the [Payload] instances defined for the given name.
func (r *PayloadRegistry[T]) Get(name string) Payloads[T] {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	namedPayloads := r.payloads[name]
	payloads := make(Payloads[T], 0, len(namedPayloads))
	payloads = append(payloads, namedPayloads...)
	return payloads
}
