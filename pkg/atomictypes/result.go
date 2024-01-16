package atomictypes

type Result struct {
	FinalCopyData CopiedData `json:"finalCopyData"`
	FinalSequence uint32     `json:"finalSequence"`
}
