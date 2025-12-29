package main

import (
	"flag"
	"fmt"
	"os"
	"seminarska/internal/client"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("/Users/izidor/Downloads/debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	addr := flag.String("addr", "", "address of the control server")
	flag.Parse()
	if *addr == "" {
		flag.Usage()
		os.Exit(1)
	}

	p := tea.NewProgram(
		client.NewAppModel(*addr),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("App exited: %v", err)
		os.Exit(1)
	}
}
