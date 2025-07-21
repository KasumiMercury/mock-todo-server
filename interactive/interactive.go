package interactive

import (
	exportHandler "github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/flagmanager"
	"github.com/KasumiMercury/mock-todo-server/server"
	"log"
)

func Start() {
	for {
		selectedCommand := commandSelector()

		switch selectedCommand {
		case serve:
			config := serveForm()

			if err := server.Run(config); err != nil {
				log.Fatal("Failed to start server:", err)
			}

			flagConfig := flagmanager.NewServeFlagConfig()
			flagConfig.FromServerConfig(config)
			flags := flagConfig.ReconstructFlags()
			displayOneLiner([]string{"serve"}, flags)
		case stop:
			err := server.Stop()
			if err != nil {
				log.Fatal("Failed to stop server:", err)
			}
			log.Println("Server stop request sent")
		case export:
			exportConfig := exportForm()
			if err := exportHandler.ExportWithMode(exportConfig.Mode, exportConfig.FilePath); err != nil {
				log.Printf("Failed to export %s: %v", exportConfig.Mode, err)
			}

			commands := []string{"export", string(exportConfig.Mode)}
			if exportConfig.FilePath != "" {
				commands = append(commands, exportConfig.FilePath)
			}

			displayOneLiner(commands, nil)
		case exit:
			return
		default:
			panic("unknown command")
		}
	}
}
