package symlinktest

import (
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/stretchr/testify/mock"
)

// MockLinkManager ...
type MockLinkManager struct {
	mock.Mock
}

// New ...
func (m *MockLinkManager) New(target, link string) *symlink.Symlink {
	args := m.Called(target, link)
	return args.Get(0).(*symlink.Symlink)
}

// Remove ...
func (m *MockLinkManager) Remove(link, target string) (*symlink.Symlink, error) {
	args := m.Called(link)
	return args.Get(0).(*symlink.Symlink), args.Error(1)
}

// Ensure ...
func (m *MockLinkManager) Ensure(link *symlink.Symlink) error {
	args := m.Called(link)
	return args.Error(0)
}

// Expand ...
func (m *MockLinkManager) Expand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

// Unexpand ...
func (m *MockLinkManager) Unexpand(link symlink.Symlink) *symlink.Symlink {
	return &link
}
