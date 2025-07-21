package interactive

import exportHandler "github.com/KasumiMercury/mock-todo-server/export"

type command int

const (
	_ command = iota
	serve
	stop
	export
	exit
)

type ExportConfig struct {
	Mode     exportHandler.ExportMode
	FilePath string
}
