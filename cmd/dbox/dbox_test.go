package main

import (
	"os"
	"testing"
)

// tempDir はテスト用の一時ディレクトリを作成する
func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dbox-test-*")
	if err != nil {
		t.Fatalf("テンポラリディレクトリ作成に失敗: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}
