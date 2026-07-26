// Package task provides a serialized task manager for Bubble Tea.
//
// TaskManager queues write operations and runs them one at a time,
// streaming output lines as tea.Msg values. It replaces all
// ad-hoc goroutines and program.Send calls in the GUI layer.
//
// See DESIGN.md (Concurrency ADR) for architecture decisions.
package task
