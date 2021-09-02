/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package status

type BatchStatus int

const (
	Unknown BatchStatus = iota
	Started
	SendCompleted
	Completed
	Failed
	Terminated
)

func (s BatchStatus) String() string {
	return [...]string{"unknown", "started", "sendCompleted", "completed", "failed", "terminated"}[s]
}
