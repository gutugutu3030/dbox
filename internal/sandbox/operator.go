package sandbox

import "time"

// SandboxOperator はサンドボックス操作を抽象化するインターフェース。
// cmd/dbox パッケージが具象型に依存せずテスト可能になるよう定義する
type SandboxOperator interface {
	FindByName(name string) (*ListOutput, error)
	Create(params CreateParams) (string, error)
	Stop(name string) error
	Remove(name string) error
	Run(name string) error
	Exec(sandboxName, command string) (string, error)
	WaitForExec(name string, timeout time.Duration) error
	PortPublish(sandboxName, portSpec string) error
	PolicyAllowNetwork(sandboxName string, domains []string) error
	List() ([]ListOutput, error)
	HasTemplate(name string) (bool, error)
	TemplateSave(tag string) error
}
