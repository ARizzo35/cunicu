// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"cunicu.li/cunicu/pkg/signaling"
	"cunicu.li/cunicu/pkg/types"
)

type Topic struct {
	subs *types.FanOut[*signaling.Envelope]
}

func NewTopic() *Topic {
	return &Topic{
		// TODO: Make smaller again
		subs: types.NewFanOut[*signaling.Envelope](10000),
	}
}

func (t *Topic) Publish(env *signaling.Envelope) {
	t.subs.Send(env)
}

func (t *Topic) Subscribe() chan *signaling.Envelope {
	return t.subs.Add()
}

func (t *Topic) Unsubscribe(ch chan *signaling.Envelope) {
	t.subs.Remove(ch)
}

func (t *Topic) Close() {
	t.subs.Close()
}
