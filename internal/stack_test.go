package internal_test

import (
	"github.com/kassett/tfstate-transfer/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStack(t *testing.T) {
	// Instantiate new stack
	stack := internal.NewStack()
	assert.NotNil(t, stack)
}

func TestStack_Push(t *testing.T) {
	stack := internal.NewStack()

	for _, element := range []string{"one", "two"} {
		stack.Push(element)
	}

	// Ensures that its LIFO
	firstElement, _ := stack.Pop()
	assert.Equal(t, firstElement, "two")
}

func TestStack_IsEmpty(t *testing.T) {
	stack := internal.NewStack()

	empty := stack.IsEmpty()
	assert.True(t, empty)

	stack.Push("One")

	empty = stack.IsEmpty()
	assert.False(t, empty)
}

func TestStack_Pop(t *testing.T) {
	stack := internal.NewStack()
	stack.Push("One")

	element, _ := stack.Pop()

	assert.Equal(t, element, "One")

	_, err := stack.Pop()
	assert.NotNil(t, err)
}
