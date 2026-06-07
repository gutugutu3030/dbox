package main

import (
	"testing"

	"github.com/gutugutu3030/sbx-template/internal/config"
	"github.com/gutugutu3030/sbx-template/internal/sandbox"
)

var testProjectCfg = &config.ProjectConfig{
	Agent:       "opencode",
	Template:    "dbox-base",
	SandboxName: "dbox-opencode-test",
	Clone:       true,
}

// TestDoStart_SandboxNotExists はサンドボックス未作成の状態で doStart を実行すると
// 作成→ポート公開→ポリシー適用→Run の順に呼ばれることを確認する
func TestDoStart_SandboxNotExists(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return nil, nil
		},
		CreateFunc: func(params sandbox.CreateParams) (string, error) {
			if params.Name != testProjectCfg.SandboxName {
				t.Errorf("params.Name = %q, want %q", params.Name, testProjectCfg.SandboxName)
			}
			if params.Agent != testProjectCfg.Agent {
				t.Errorf("params.Agent = %q, want %q", params.Agent, testProjectCfg.Agent)
			}
			if params.Template != testProjectCfg.Template {
				t.Errorf("params.Template = %q, want %q", params.Template, testProjectCfg.Template)
			}
			return "", nil
		},
	}

	if err := doStart(mock, testProjectCfg.SandboxName, []string{}, testProjectCfg); err != nil {
		t.Fatalf("doStart() エラー: %v", err)
	}

	// 呼び出し順序の検証（空ポートのため PortPublish は呼ばれない）
	expectedMethods := []string{"FindByName", "Create", "WaitForExec", "PolicyAllowNetwork", "Run"}
	if len(mock.Calls) < len(expectedMethods) {
		t.Fatalf("呼び出し回数 = %d, want >= %d", len(mock.Calls), len(expectedMethods))
	}
	for i, expected := range expectedMethods {
		if mock.Calls[i].Method != expected {
			t.Errorf("Call[%d] = %q, want %q", i, mock.Calls[i].Method, expected)
		}
	}
}

// TestDoStart_SandboxStopped は停止中のサンドボックスで doStart を実行すると
// ポリシー適用→Run の順に呼ばれることを確認する
func TestDoStart_SandboxStopped(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return &sandbox.ListOutput{
				Sandbox: name,
				Status:  "stopped",
				Agent:   "opencode",
			}, nil
		},
	}

	if err := doStart(mock, "dbox-opencode-test", []string{}, testProjectCfg); err != nil {
		t.Fatalf("doStart() エラー: %v", err)
	}

	// 停止中 → PolicyAllowNetwork → Run
	expectedMethods := []string{"FindByName", "PolicyAllowNetwork", "Run"}
	if len(mock.Calls) < len(expectedMethods) {
		t.Fatalf("呼び出し回数 = %d, want >= %d", len(mock.Calls), len(expectedMethods))
	}
	for i, expected := range expectedMethods {
		if mock.Calls[i].Method != expected {
			t.Errorf("Call[%d] = %q, want %q", i, mock.Calls[i].Method, expected)
		}
	}

	// Create が呼ばれていないこと
	if mock.AssertCalled("Create") {
		t.Error("停止中の場合に Create が呼ばれています")
	}
}

// TestDoStart_SandboxRunning は起動中のサンドボックスで doStart を実行すると
// WaitForExec→PortPublish→PolicyAllowNetwork→Run の順に呼ばれることを確認する
func TestDoStart_SandboxRunning(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return &sandbox.ListOutput{
				Sandbox: name,
				Status:  "running",
				Agent:   "opencode",
			}, nil
		},
	}

	if err := doStart(mock, "dbox-opencode-test", []string{"3000"}, testProjectCfg); err != nil {
		t.Fatalf("doStart() エラー: %v", err)
	}

	// 起動中 → WaitForExec → PortPublish → PolicyAllowNetwork → Run
	expectedMethods := []string{"FindByName", "WaitForExec", "PortPublish", "PolicyAllowNetwork", "Run"}
	if len(mock.Calls) < len(expectedMethods) {
		t.Fatalf("呼び出し回数 = %d, want >= %d", len(mock.Calls), len(expectedMethods))
	}
	for i, expected := range expectedMethods {
		if mock.Calls[i].Method != expected {
			t.Errorf("Call[%d] = %q, want %q", i, mock.Calls[i].Method, expected)
		}
	}

	// Create が呼ばれていないこと
	if mock.AssertCalled("Create") {
		t.Error("起動中の場合に Create が呼ばれています")
	}
}

// TestDoStart_SandboxNotExists_WithPorts は未作成時にポート指定があれば PortPublish が呼ばれることを確認する
func TestDoStart_SandboxNotExists_WithPorts(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return nil, nil
		},
		CreateFunc: func(params sandbox.CreateParams) (string, error) {
			return "", nil
		},
	}

	if err := doStart(mock, "dbox-opencode-test", []string{"8080", "3000"}, testProjectCfg); err != nil {
		t.Fatalf("doStart() エラー: %v", err)
	}

	if mock.CallCount("PortPublish") != 2 {
		t.Errorf("PortPublish の呼び出し回数 = %d, want 2", mock.CallCount("PortPublish"))
	}
	if mock.CallCount("PolicyAllowNetwork") != 1 {
		t.Errorf("PolicyAllowNetwork の呼び出し回数 = %d, want 1", mock.CallCount("PolicyAllowNetwork"))
	}
	if mock.CallCount("Run") != 1 {
		t.Errorf("Run の呼び出し回数 = %d, want 1", mock.CallCount("Run"))
	}
}

// TestDoStart_FindByNameError は FindByName のエラーが伝播されることを確認する
func TestDoStart_FindByNameError(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return nil, assertAnError("find error")
		},
	}

	err := doStart(mock, "dbox-opencode-test", []string{}, testProjectCfg)
	if err == nil {
		t.Fatal("doStart() はエラーを返すべきです")
	}
}

// TestDoStart_CreateError は Create のエラーが伝播されることを確認する
func TestDoStart_CreateError(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return nil, nil
		},
		CreateFunc: func(params sandbox.CreateParams) (string, error) {
			return "", assertAnError("create error")
		},
	}

	err := doStart(mock, "dbox-opencode-test", []string{}, testProjectCfg)
	if err == nil {
		t.Fatal("doStart() はエラーを返すべきです")
	}
}

// TestDoStart_ProjectCfgNil は projectCfg が nil の場合にエラーになることを確認する
func TestDoStart_ProjectCfgNil(t *testing.T) {
	mock := &MockSandboxOperator{
		FindByNameFunc: func(name string) (*sandbox.ListOutput, error) {
			return nil, nil
		},
	}

	err := doStart(mock, "dbox-opencode-test", []string{}, nil)
	if err == nil {
		t.Fatal("doStart() は projectCfg が nil の場合エラーを返すべきです")
	}
}
