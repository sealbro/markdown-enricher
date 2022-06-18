package closer

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Closer interface {
	Close(ctx context.Context) error
}

type CloserCollection struct {
	services map[string]Closer
	m        sync.Mutex
}

func MakeCloserCollection() *CloserCollection {
	return &CloserCollection{
		services: make(map[string]Closer),
		m:        sync.Mutex{},
	}
}

func (c *CloserCollection) Add(closer Closer) {
	c.m.Lock()
	c.services[reflect.TypeOf(closer).String()] = closer
	c.m.Unlock()
}

func (c *CloserCollection) Close(ctx context.Context) error {
	builder := strings.Builder{}
	for _, closer := range c.services {
		err := closer.Close(ctx)
		if err != nil {
			builder.WriteString(fmt.Sprintf("%v\n", err))
		}
	}

	s := builder.String()
	if len(s) > 0 {
		return fmt.Errorf(s)
	}

	return nil
}
