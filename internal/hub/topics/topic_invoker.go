package topics

import "context"

type TopicInvoker interface {
	InvokeTopic(ctx context.Context, payload []byte) error
}
