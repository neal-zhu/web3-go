package atomicals

const MAX_SEQUENCE = 0xffffffff

type Result struct {
	FinalCopyData CopiedData `json:"finalCopyData"`
	FinalSequence uint32     `json:"finalSequence"`
}
