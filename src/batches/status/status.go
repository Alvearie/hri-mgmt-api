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
	Completed
	Failed
	Terminated
)

func (s BatchStatus) String() string {
	return [...]string{"unknown", "started", "completed", "failed", "terminated"}[s]
}
