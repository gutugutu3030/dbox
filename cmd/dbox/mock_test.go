package main

import (
	"fmt"
	"time"

	"github.com/gutugutu3030/sbx-template/internal/sandbox"
)

// CallRecord はモックメソッドの呼び出しを記録する
type CallRecord struct {
	Method string
	Args   []any
}

// MockSandboxOperator は SandboxOperator のモック実装。
// 呼び出し履歴を記録し、任意の返り値を設定できる
type MockSandboxOperator struct {
	// 呼び出し履歴
	Calls []CallRecord

	// 各メソッドの返り値を制御する関数（デフォルトはゼロ値を返す）
	FindByNameFunc   func(name string) (*sandbox.ListOutput, error)
	CreateFunc       func(params sandbox.CreateParams) (string, error)
	StopFunc         func(name string) error
	RemoveFunc       func(name string) error
	RunFunc          func(name string) error
	ExecFunc         func(sandboxName, command string) (string, error)
	WaitForExecFunc  func(name string, timeout time.Duration) error
	PortPublishFunc  func(sandboxName, portSpec string) error
	PolicyAllowFunc  func(sandboxName string, domains []string) error
	ListFunc         func() ([]sandbox.ListOutput, error)
	HasTemplateFunc  func(name string) (bool, error)
	TemplateSaveFunc func(tag string) error
}

func (m *MockSandboxOperator) record(method string, args ...any) {
	m.Calls = append(m.Calls, CallRecord{Method: method, Args: args})
}

// AssertCalled は指定されたメソッドが指定回数呼ばれたことを検証する
func (m *MockSandboxOperator) AssertCalled(method string) bool {
	for _, c := range m.Calls {
		if c.Method == method {
			return true
		}
	}
	return false
}

// AssertNotCalled は指定されたメソッドが一度も呼ばれていないことを検証する
func (m *MockSandboxOperator) AssertNotCalled(method string) bool {
	for _, c := range m.Calls {
		if c.Method == method {
			return false
		}
	}
	return true
}

// CallCount は指定されたメソッドの呼び出し回数を返す
func (m *MockSandboxOperator) CallCount(method string) int {
	count := 0
	for _, c := range m.Calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

func (m *MockSandboxOperator) FindByName(name string) (*sandbox.ListOutput, error) {
	m.record("FindByName", name)
	if m.FindByNameFunc != nil {
		return m.FindByNameFunc(name)
	}
	return nil, nil
}

func (m *MockSandboxOperator) Create(params sandbox.CreateParams) (string, error) {
	m.record("Create", params)
	if m.CreateFunc != nil {
		return m.CreateFunc(params)
	}
	return "", nil
}

func (m *MockSandboxOperator) Stop(name string) error {
	m.record("Stop", name)
	if m.StopFunc != nil {
		return m.StopFunc(name)
	}
	return nil
}

func (m *MockSandboxOperator) Remove(name string) error {
	m.record("Remove", name)
	if m.RemoveFunc != nil {
		return m.RemoveFunc(name)
	}
	return nil
}

func (m *MockSandboxOperator) Run(name string) error {
	m.record("Run", name)
	if m.RunFunc != nil {
		return m.RunFunc(name)
	}
	return nil
}

func (m *MockSandboxOperator) Exec(sandboxName, command string) (string, error) {
	m.record("Exec", sandboxName, command)
	if m.ExecFunc != nil {
		return m.ExecFunc(sandboxName, command)
	}
	return "", nil
}

func (m *MockSandboxOperator) WaitForExec(name string, timeout time.Duration) error {
	m.record("WaitForExec", name, timeout)
	if m.WaitForExecFunc != nil {
		return m.WaitForExecFunc(name, timeout)
	}
	return nil
}

func (m *MockSandboxOperator) PortPublish(sandboxName, portSpec string) error {
	m.record("PortPublish", sandboxName, portSpec)
	if m.PortPublishFunc != nil {
		return m.PortPublishFunc(sandboxName, portSpec)
	}
	return nil
}

func (m *MockSandboxOperator) PolicyAllowNetwork(sandboxName string, domains []string) error {
	m.record("PolicyAllowNetwork", sandboxName, domains)
	if m.PolicyAllowFunc != nil {
		return m.PolicyAllowFunc(sandboxName, domains)
	}
	return nil
}

func (m *MockSandboxOperator) List() ([]sandbox.ListOutput, error) {
	m.record("List")
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

func (m *MockSandboxOperator) HasTemplate(name string) (bool, error) {
	m.record("HasTemplate", name)
	if m.HasTemplateFunc != nil {
		return m.HasTemplateFunc(name)
	}
	return false, nil
}

func (m *MockSandboxOperator) TemplateSave(tag string) error {
	m.record("TemplateSave", tag)
	if m.TemplateSaveFunc != nil {
		return m.TemplateSaveFunc(tag)
	}
	return nil
}

// String はモックの呼び出し履歴を文字列で返す（デバッグ用）
func (m *MockSandboxOperator) String() string {
	s := "MockSandboxOperator calls:\n"
	for _, c := range m.Calls {
		s += fmt.Sprintf("  %s(%v)\n", c.Method, c.Args)
	}
	return s
}
