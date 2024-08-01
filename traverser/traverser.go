package traverser

import "context"

type Traverser interface {
	Initialize(ctx context.Context)
	Traverse(ctx context.Context) (chan map[string][]string, chan map[string][]string)
}
