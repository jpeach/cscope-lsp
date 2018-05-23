package cquery

// Progress is a message sent by the cquery "$cquery/progress" notification.
type Progress struct {
	IndexRequestCount      int `json:"indexRequestCount"`
	DoIDMapCount           int `json:"doIdMapCount"`
	LoadPreviousIndexCount int `json:"loadPreviousIndexCount"`
	OnIDMappedCount        int `json:"onIdMappedCount"`
	OnIndexedCount         int `json:"onIndexedCount"`
	ActiveThreads          int `json:"activeThreads"`
}
