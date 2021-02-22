package model

type TriggerType uint
type SourceValueType uint

const (
	ExponentialMovingAverages SourceValueType = iota
	SimpleMovingAverages
	MACD
	RSI
)

const (
	Range TriggerType = iota
	LessThan
	MoreThan
)

type Trigger struct {
	ID           uint            `json:"id"`
	Name         string          `json:"name"`
	Type         TriggerType     `json:"type"`
	AnalysisType SourceValueType `json:"source_type"`
}
