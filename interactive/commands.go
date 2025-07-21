package interactive

import "github.com/charmbracelet/huh"

func commandSelector() command {
	var selectedCommand command
	selector := huh.NewSelect[command]().
		Title("Select a command").
		Options(
			huh.NewOption("Start Server", serve),
			huh.NewOption("Stop Server", stop),
			huh.NewOption("Export JSON Data/Template", export),
			huh.NewOption("Exit", exit),
		).
		Value(&selectedCommand)

	if err := selector.Run(); err != nil {
		panic("failed to select command: " + err.Error())
	}

	return selectedCommand
}
