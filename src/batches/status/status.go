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

func GetBatchStatus(statusStr string) BatchStatus {
	switch statusStr {
	case Started.String():
		return Started
	case SendCompleted.String():
		return SendCompleted
	case Completed.String():
		return Completed
	case Failed.String():
		return Failed
	case Terminated.String():
		return Terminated
	}

	return Unknown
}
