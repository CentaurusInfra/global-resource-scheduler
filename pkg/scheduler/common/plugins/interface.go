package plugins

import (
	"errors"
	"math"
	"strings"
)

// NodeScoreList declares a list of nodes and their scores.
type NodeScoreList []NodeScore

// NodeScore is a struct with node name and score.
type NodeScore struct {
	Name  string
	Score int64
}

// PluginToNodeScores declares a map from plugin name to its NodeScoreList.
type PluginToNodeScores map[string]NodeScoreList

// NodeToStatusMap declares map from node name to its status.
type NodeToStatusMap map[string]*Status

// Code is the Status code/type which is returned from plugins.
type Code int

// These are predefined codes used in a Status.
const (
	// Success means that plugin ran correctly and found pod schedulable.
	// NOTE: A nil status is also considered as "Success".
	Success Code = iota
	// Error is used for internal plugin errors, unexpected input, etc.
	Error
	// Unschedulable is used when a plugin finds a pod unschedulable. The scheduler might attempt to
	// preempt other pods to get this pod scheduled. Use UnschedulableAndUnresolvable to make the
	// scheduler skip preemption.
	// The accompanying status message should explain why the pod is unschedulable.
	Unschedulable
	// UnschedulableAndUnresolvable is used when a (pre-)filter plugin finds a pod unschedulable and
	// preemption would not change anything. Plugins should return Unschedulable if it is possible
	// that the pod can get scheduled with preemption.
	// The accompanying status message should explain why the pod is unschedulable.
	UnschedulableAndUnresolvable
	// Wait is used when a permit plugin finds a pod scheduling should wait.
	Wait
	// Skip is used when a bind plugin chooses to skip binding.
	Skip
)

// This list should be exactly the same as the codes iota defined above in the same order.
var codes = []string{"Success", "Error", "Unschedulable", "UnschedulableAndUnresolvable", "Wait", "Skip"}

func (c Code) String() string {
	return codes[c]
}

const (
	// MaxNodeScore is the maximum score a Score plugin is expected to return.
	MaxNodeScore int64 = 100

	// MinNodeScore is the minimum score a Score plugin is expected to return.
	MinNodeScore int64 = 0

	// MaxTotalScore is the maximum total score.
	MaxTotalScore int64 = math.MaxInt64
)

// Status indicates the result of running a plugin. It consists of a code and a
// message. When the status code is not `Success`, the reasons should explain why.
// NOTE: A nil Status is also considered as Success.
type Status struct {
	code    Code
	reasons []string
}

// Code returns code of the Status.
func (s *Status) Code() Code {
	if s == nil {
		return Success
	}
	return s.code
}

// Message returns a concatenated message on reasons of the Status.
func (s *Status) Message() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.reasons, ", ")
}

// Reasons returns reasons of the Status.
func (s *Status) Reasons() []string {
	return s.reasons
}

// AppendReason appends given reason to the Status.
func (s *Status) AppendReason(reason string) {
	s.reasons = append(s.reasons, reason)
}

// IsSuccess returns true if and only if "Status" is nil or Code is "Success".
func (s *Status) IsSuccess() bool {
	return s.Code() == Success
}

// IsUnschedulable returns true if "Status" is Unschedulable (Unschedulable or UnschedulableAndUnresolvable).
func (s *Status) IsUnschedulable() bool {
	code := s.Code()
	return code == Unschedulable || code == UnschedulableAndUnresolvable
}

// AsError returns nil if the status is a success; otherwise returns an "error" object
// with a concatenated message on reasons of the Status.
func (s *Status) AsError() error {
	if s.IsSuccess() {
		return nil
	}
	return errors.New(s.Message())
}

// NewStatus makes a Status out of the given arguments and returns its pointer.
func NewStatus(code Code, reasons ...string) *Status {
	return &Status{
		code:    code,
		reasons: reasons,
	}
}

// PluginToStatus maps plugin name to status. Currently used to identify which Filter plugin
// returned which status.
type PluginToStatus map[string]*Status

// Merge merges the statuses in the map into one. The resulting status code have the following
// precedence: Error, UnschedulableAndUnresolvable, Unschedulable.
func (p PluginToStatus) Merge() *Status {
	if len(p) == 0 {
		return nil
	}

	finalStatus := NewStatus(Success)
	var hasError, hasUnschedulableAndUnresolvable, hasUnschedulable bool
	for _, s := range p {
		if s.Code() == Error {
			hasError = true
		} else if s.Code() == UnschedulableAndUnresolvable {
			hasUnschedulableAndUnresolvable = true
		} else if s.Code() == Unschedulable {
			hasUnschedulable = true
		}
		finalStatus.code = s.Code()
		for _, r := range s.reasons {
			finalStatus.AppendReason(r)
		}
	}

	if hasError {
		finalStatus.code = Error
	} else if hasUnschedulableAndUnresolvable {
		finalStatus.code = UnschedulableAndUnresolvable
	} else if hasUnschedulable {
		finalStatus.code = Unschedulable
	}
	return finalStatus
}

// Plugin is the parent type for all the scheduling framework plugins.
type Plugin interface {
	Name() string
}
