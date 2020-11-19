package ldevents

type nullEventProcessor struct{}

// NewNullEventProcessor creates a no-op implementation of EventProcessor.
func NewNullEventProcessor() EventProcessor {
	return nullEventProcessor{}
}

func (n nullEventProcessor) RecordFeatureRequestEvent(e FeatureRequestEvent) {}

func (n nullEventProcessor) RecordIdentifyEvent(e IdentifyEvent) {}

func (n nullEventProcessor) RecordCustomEvent(e CustomEvent) {}

func (n nullEventProcessor) Flush() {}

func (n nullEventProcessor) Close() error {
	return nil
}
