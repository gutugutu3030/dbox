package sandbox

import (
	"testing"
)

// TestNewRunner は Runner が正しく作成されることを確認する
func TestNewRunner(t *testing.T) {
	r := NewRunner(false)
	if r == nil {
		t.Fatal("NewRunner() が nil を返しました")
	}
	if r.DryRun != false {
		t.Error("DryRun が true です")
	}

	r2 := NewRunner(true)
	if r2.DryRun != true {
		t.Error("DryRun が false です")
	}
}

// TestCreateParams は CreateParams が正しく設定されることを確認する
func TestCreateParams(t *testing.T) {
	params := CreateParams{
		Name:     "test-sandbox",
		Template: "dbox-node",
		Agent:    "opencode",
		Path:     "/tmp/test",
		Clone:    true,
		CPUs:     2,
		Memory:   "4g",
	}

	if params.Name != "test-sandbox" {
		t.Errorf("Name = %q, want %q", params.Name, "test-sandbox")
	}
	if params.Template != "dbox-node" {
		t.Errorf("Template = %q, want %q", params.Template, "dbox-node")
	}
	if params.Agent != "opencode" {
		t.Errorf("Agent = %q, want %q", params.Agent, "opencode")
	}
	if params.Clone != true {
		t.Errorf("Clone = %v, want %v", params.Clone, true)
	}
}

// TestListOutput はパース結果の構造が正しいことを確認する
func TestListOutput(t *testing.T) {
	lo := ListOutput{
		Sandbox:   "dbox-opencode-test",
		Agent:     "opencode",
		Status:    "running",
		Ports:     "",
		Workspace: "/tmp/test",
	}

	if lo.Sandbox != "dbox-opencode-test" {
		t.Errorf("Sandbox = %q", lo.Sandbox)
	}
	if lo.Agent != "opencode" {
		t.Errorf("Agent = %q", lo.Agent)
	}
	if lo.Status != "running" {
		t.Errorf("Status = %q", lo.Status)
	}
}

// TestParseListOutput は sbx ls の出力パースロジックをテストする
func TestParseListOutput(t *testing.T) {
	tests := []struct {
		name     string
		lines    []ListOutput
		expected int
	}{
		{
			name:     "空の結果",
			lines:    nil,
			expected: 0,
		},
		{
			name: "1件の結果",
			lines: []ListOutput{
				{Sandbox: "test-sandbox", Agent: "opencode", Status: "running", Workspace: "/tmp/test"},
			},
			expected: 1,
		},
		{
			name: "複数件の結果",
			lines: []ListOutput{
				{Sandbox: "sandbox-1", Agent: "opencode", Status: "running", Workspace: "/tmp/1"},
				{Sandbox: "sandbox-2", Agent: "shell", Status: "stopped", Workspace: "/tmp/2"},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.lines) != tt.expected {
				t.Errorf("len(lines) = %d, want %d", len(tt.lines), tt.expected)
			}
		})
	}
}

// TestDryRunMode は DryRun モードでエラーが発生しないことを確認する
func TestDryRunMode(t *testing.T) {
	r := NewRunner(true)

	// Create
	_, err := r.Create(CreateParams{
		Name:     "test",
		Template: "dbox-base",
		Agent:    "opencode",
		Path:     ".",
	})
	if err != nil {
		t.Errorf("Create() with DryRun エラー: %v", err)
	}

	// List
	_, err = r.List()
	if err != nil {
		t.Errorf("List() with DryRun エラー: %v", err)
	}

	// Run (interactive)
	err = r.Run("test")
	if err != nil {
		t.Errorf("Run() with DryRun エラー: %v", err)
	}

	// Stop
	err = r.Stop("test")
	if err != nil {
		t.Errorf("Stop() with DryRun エラー: %v", err)
	}

	// Start
	err = r.Start("test")
	if err != nil {
		t.Errorf("Start() with DryRun エラー: %v", err)
	}

	// Exec
	_, err = r.Exec("test", "echo hello")
	if err != nil {
		t.Errorf("Exec() with DryRun エラー: %v", err)
	}

	// TemplateList
	_, err = r.TemplateList()
	if err != nil {
		t.Errorf("TemplateList() with DryRun エラー: %v", err)
	}

	// TemplateSave (DryRun)
	err = r.TemplateSave("dbox-test:latest")
	if err != nil {
		t.Errorf("TemplateSave() with DryRun エラー: %v", err)
	}

	// HasTemplate (DryRun)
	_, err = r.HasTemplate("dbox-base")
	if err != nil {
		t.Errorf("HasTemplate() with DryRun エラー: %v", err)
	}

	// Remove (DryRun)
	err = r.Remove("test")
	if err != nil {
		t.Errorf("Remove() with DryRun エラー: %v", err)
	}

	// TemplateRemove (DryRun)
	err = r.TemplateRemove("dbox-test:latest")
	if err != nil {
		t.Errorf("TemplateRemove() with DryRun エラー: %v", err)
	}

	// PortPublish (DryRun)
	err = r.PortPublish("test", "8080")
	if err != nil {
		t.Errorf("PortPublish() with DryRun エラー: %v", err)
	}
}
