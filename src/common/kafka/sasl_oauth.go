/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go/sasl"
)

const (
	gs2_header = "n,,"
	kvsep      = "\x01"
)

// This class implements a minimal OAUTHBEARER SASL mechanism for Kafka-go
// Implementation is based off the following documentation: https://tools.ietf.org/html/rfc7628
type OAuthBearer struct {
	Token string
}

func (OAuthBearer) Name() string {
	return "OAUTHBEARER"
}

func (oath OAuthBearer) Start(ctx context.Context) (sasl.StateMachine, []byte, error) {
	return oath, []byte(fmt.Sprintf("%s%sauth=Bearer %s%s%s", gs2_header, kvsep, oath.Token, kvsep, kvsep)), nil
}

func (oath OAuthBearer) Next(ctx context.Context, challenge []byte) (bool, []byte, error) {
	if challenge != nil && len(challenge) > 0 {
		// error scenario
		fmt.Println(string(challenge))
		return false, []byte("\x01"), nil
	}
	return true, nil, nil
}
