//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

// Package notify provides a generic interface to send user notifications.
// The actual transport behind (mail, webhook) can be plugged in as needed.
package notify

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"reflect"
	"strings"
	"sync"
	texttemplate "text/template"
)

// Payload represents a single notification including content and transport
// specific attributes.
type Payload interface {
	// Send sends the Payload after applying the given params object.
	Send(ctx context.Context, params any) error
}

// Payloads represents an array of [Payload] instances.
type Payloads []Payload

// Send invokes [Payload.Send] for all given [Payload] instances.
func (payloads Payloads) Send(ctx context.Context, params any) error {
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
type PayloadRegistry struct {
	payloads map[string][]Payload
	mutex    sync.RWMutex
}

// Add adds a the given [Payload] instance using the given name. Multiple
// payloads can be defined for a single name.
func (r *PayloadRegistry) Add(name string, payload Payload) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.payloads == nil {
		r.payloads = make(map[string][]Payload)
	}
	namedPayloads := r.payloads[name]
	namedPayloads = append(namedPayloads, payload)
	r.payloads[name] = namedPayloads
}

// Get gets the [Payload] instances defined for the given name.
func (r *PayloadRegistry) Get(name string) Payloads {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	namedPayloads := r.payloads[name]
	payloads := make(Payloads, 0, len(namedPayloads))
	payloads = append(payloads, namedPayloads...)
	return payloads
}

// NamedValue represents a named value pair as found in a parameter object.
type NamedValue struct {
	Name  string
	Value any
}

// DecodeParams decodes a struct object to the contained [NamedValue] pairs.
// If the params object is nil, an empty array is returned.
func DecodeParams(params any) ([]NamedValue, error) {
	if params == nil {
		return []NamedValue{}, nil
	}
	paramsValue := reflect.ValueOf(params)
	if paramsValue.Kind() != reflect.Ptr || paramsValue.Elem().Kind() != reflect.Struct {
		return nil, errors.New("invalid params type; must be pointer to struct")
	}
	structValue := paramsValue.Elem()
	structType := structValue.Type()
	numField := structValue.NumField()
	namedValues := make([]NamedValue, numField)
	for i := range numField {
		field := structType.Field(i)
		namedValues = append(namedValues, NamedValue{Name: field.Name, Value: structValue.Field(i)})
	}
	return namedValues, nil
}

// ExecuteTextTemplate invokes [texttemplat.Execute] using the given template text
// and parameter object.
func ExecuteTextTemplate(text string, params any) (string, error) {
	if params == nil {
		return text, nil
	}
	tmpl, err := texttemplate.New("message").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template (cause: %w)", err)
	}
	buffer := &strings.Builder{}
	err = tmpl.Execute(buffer, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute text template (cause: %w)", err)
	}
	return buffer.String(), nil
}

// ExecuteHTMLTemplate invokes [htmltemplat.Execute] using the given template text
// and parameter object.
func ExecuteHTMLTemplate(html string, params any) (string, error) {
	if params == nil {
		return html, nil
	}
	tmpl, err := htmltemplate.New("message").Parse(html)
	if err != nil {
		return "", fmt.Errorf("failed to parse html template (cause: %w)", err)
	}
	buffer := &strings.Builder{}
	err = tmpl.Execute(buffer, params)
	if err != nil {
		return "", fmt.Errorf("failed to execute html template (cause: %w)", err)
	}
	return buffer.String(), nil
}
