package prompt

import (
	"testing"
)

// TestExtractNameFromPackage はパッケージ名からサーバー名を抽出することを確認する
func TestExtractNameFromPackage(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{pkg: "@anthropic/gitnexus", want: "gitnexus"},
		{pkg: "@scope/context7-mcp", want: "context7-mcp"},
		{pkg: "simple-package", want: "simple-package"},
		{pkg: "@a/b/c", want: "c"}, // 最終セグメントを抽出
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := extractNameFromPackage(tt.pkg)
			if got != tt.want {
				t.Errorf("extractNameFromPackage(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}

// TestIsInteractive は非 CI 環境で IsInteractive が呼べることを確認する
func TestIsInteractive(t *testing.T) {
	// CI 環境では false になることを確認
	t.Setenv("CI", "true")
	if IsInteractive() {
		t.Error("CI=true で IsInteractive() が true を返しました")
	}
}
