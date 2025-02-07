package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/appclacks/maizai/pkg/shared"
	er "github.com/mcorbin/corbierror"
)

type MemoryContextStore struct {
	state map[string]*shared.Context
	lock  sync.RWMutex
}

func New() *MemoryContextStore {
	return &MemoryContextStore{
		state: make(map[string]*shared.Context),
	}
}

func (m *MemoryContextStore) DeleteContext(ctx context.Context, id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.state, id)
	return nil
}

func (m *MemoryContextStore) ContextExists(ctx context.Context, id string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.state[id]
	return ok, nil
}

func (m *MemoryContextStore) ContextExistsByName(ctx context.Context, name string) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, context := range m.state {
		if context.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *MemoryContextStore) get(_ context.Context, id string) (*shared.Context, error) {
	context, ok := m.state[id]
	if !ok {
		return nil, er.Newf("context %s doesn't exist", er.NotFound, true, id)
	}
	return context, nil
}

func (m *MemoryContextStore) GetContext(ctx context.Context, id string) (*shared.Context, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.get(ctx, id)
}

func (m *MemoryContextStore) GetByName(ctx context.Context, name string) (*shared.Context, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, context := range m.state {
		if context.Name == name {
			return context, nil
		}
	}
	return nil, fmt.Errorf("context %s doesn't exist", name)
}

func (m *MemoryContextStore) CreateContext(ctx context.Context, context shared.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.state[context.ID] = &context
	return nil
}

func (m *MemoryContextStore) ListContexts(ctx context.Context) ([]shared.ContextMetadata, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result := []shared.ContextMetadata{}
	for k, v := range m.state {
		result = append(result, shared.ContextMetadata{
			ID:          k,
			Name:        v.Name,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			Sources:     v.Sources,
		})
	}
	return result, nil
}

func (m *MemoryContextStore) AddMessages(ctx context.Context, id string, messages []shared.Message) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	context, err := m.get(ctx, id)
	if err != nil {
		return err
	}
	context.Messages = append(context.Messages, messages...)
	return nil
}

func (m *MemoryContextStore) DeleteContextMessage(ctx context.Context, messageID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, context := range m.state {
		for i := range context.Messages {
			msgID := context.Messages[i].ID
			if msgID == messageID {
				context.Messages = append(context.Messages[:i], context.Messages[i+1:]...)
				return nil

			}
		}
	}
	return nil
}

func (m *MemoryContextStore) DeleteContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	context, ok := m.state[contextID]
	if !ok {
		return fmt.Errorf("context %s doesn't exist", contextID)
	}
	sources := []string{}
	found := false
	for _, source := range context.Sources.Contexts {
		if source != sourceContextID {
			sources = append(sources, source)
		} else {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("context source %s doesn't exist", sourceContextID)
	}
	context.Sources.Contexts = sources
	return nil
}

func (m *MemoryContextStore) UpdateContextMessage(ctx context.Context, messageID string, role string, content string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, context := range m.state {
		for i := range context.Messages {
			msgID := context.Messages[i].ID
			if msgID == messageID {
				context.Messages[i].Content = content
				context.Messages[i].Role = role
				return nil
			}
		}
	}
	return nil
}

func (m *MemoryContextStore) CreateContextSourceContext(ctx context.Context, contextID string, sourceContextID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	context, ok := m.state[contextID]
	if !ok {
		return fmt.Errorf("context %s doesn't exist", sourceContextID)
	}
	context.Sources.Contexts = append(context.Sources.Contexts, sourceContextID)
	return nil
}
