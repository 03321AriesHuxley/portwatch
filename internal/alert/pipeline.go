package alert

import (
	"context"
	"fmt"
	"strings"
)

// PipelineStage represents a named step in a notifier pipeline.
type PipelineStage struct {
	Name     string
	Notifier Notifier
}

// Pipeline chains multiple notifier stages together in order, passing events
// through each stage sequentially. If any stage returns an error the pipeline
// halts and returns that error annotated with the stage name.
//
// Example:
//
//	pipeline := NewPipeline(
//		PipelineStage{"filter",   NewFilter(notifier, opts)},
//		PipelineStage{"throttle", NewThrottleNotifier(notifier, window)},
//		PipelineStage{"batch",    NewBatchNotifier(notifier, size, window)},
//	)
type Pipeline struct {
	stages []PipelineStage
}

// NewPipeline constructs a Pipeline from an ordered list of stages.
// Stages are executed in the order they are provided.
func NewPipeline(stages ...PipelineStage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Send passes events through each stage in order.
// Execution stops at the first stage that returns an error.
func (p *Pipeline) Send(ctx context.Context, events []Event) error {
	for _, stage := range p.stages {
		if err := stage.Notifier.Send(ctx, events); err != nil {
			return fmt.Errorf("pipeline stage %q: %w", stage.Name, err)
		}
	}
	return nil
}

// Stages returns a copy of the stage slice for inspection.
func (p *Pipeline) Stages() []PipelineStage {
	out := make([]PipelineStage, len(p.stages))
	copy(out, p.stages)
	return out
}

// Len returns the number of stages in the pipeline.
func (p *Pipeline) Len() int {
	return len(p.stages)
}

// String returns a human-readable description of the pipeline stages.
func (p *Pipeline) String() string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.Name
	}
	return "Pipeline[" + strings.Join(names, " -> ") + "]"
}

// PipelineBuilder provides a fluent API for constructing a Pipeline.
//
//	pipeline := NewPipelineBuilder().
//		Add("dedup",    NewDeduplicator(inner, window)).
//		Add("throttle", NewThrottleNotifier(inner, window)).
//		Build()
type PipelineBuilder struct {
	stages []PipelineStage
}

// NewPipelineBuilder returns an empty PipelineBuilder.
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{}
}

// Add appends a named stage to the builder.
func (b *PipelineBuilder) Add(name string, n Notifier) *PipelineBuilder {
	b.stages = append(b.stages, PipelineStage{Name: name, Notifier: n})
	return b
}

// Build returns the assembled Pipeline.
func (b *PipelineBuilder) Build() *Pipeline {
	return NewPipeline(b.stages...)
}
