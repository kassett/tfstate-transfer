package internal

import "errors"

type Stack struct {
	elements []string
}

// NewStack creates a new stack
func NewStack() *Stack {
	return &Stack{elements: []string{}}
}

// Push adds an element to the stack
func (s *Stack) Push(element string) {
	s.elements = append(s.elements, element)
}

// Pop removes and returns the top element of the stack
func (s *Stack) Pop() (string, error) {
	if len(s.elements) == 0 {
		return "", errors.New("stack is empty")
	}
	topIndex := len(s.elements) - 1
	topElement := s.elements[topIndex]
	s.elements = s.elements[:topIndex]
	return topElement, nil
}

func (s *Stack) IsEmpty() bool {
	return len(s.elements) == 0
}
