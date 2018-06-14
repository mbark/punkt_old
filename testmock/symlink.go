package testmock

import (
	"github.com/mbark/punkt/pkg/mgr/symlink"
	"github.com/stretchr/testify/mock"
)

// LinkManager ...
type LinkManager struct {
	mock.Mock
}

// New ...
func (m *LinkManager) New(target, link string) *symlink.Symlink {
	args := m.Called(target, link)
	return args.Get(0).(*symlink.Symlink)
}

// Remove ...
func (m *LinkManager) Remove(link string) (*symlink.Symlink, error) {
	args := m.Called(link)
	return args.Get(0).(*symlink.Symlink), args.Error(1)
}

// Ensure ...
func (m *LinkManager) Ensure(link *symlink.Symlink) error {
	args := m.Called(link)
	return args.Error(0)
}

// Expand ...
func (m *LinkManager) Expand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

// Unexpand ...
func (m *LinkManager) Unexpand(link symlink.Symlink) *symlink.Symlink {
	return &link
}
