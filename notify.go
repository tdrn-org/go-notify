//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package notify

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

type Payload interface {
	Send(ctx context.Context, params any) error
}

type Payloads []Payload

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

type PayloadRegistry struct {
	payloads map[string][]Payload
	mutex    sync.RWMutex
}

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

func (r *PayloadRegistry) Get(name string) Payloads {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	namedPayloads := r.payloads[name]
	payloads := make(Payloads, 0, len(namedPayloads))
	payloads = append(payloads, namedPayloads...)
	return payloads
}

type NamedPayloadFactory interface {
	Create(name string) (Payload, error)
}

type NamedValue struct {
	Name  string
	Value any
}

func DecodeParams(params any) ([]NamedValue, error) {
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
		tag := field.Tag.Get("name")
		if tag != "" && tag != "-" {
			namedValues = append(namedValues, NamedValue{Name: tag, Value: structValue.Field(i)})
		}
	}
	return namedValues, nil
}
