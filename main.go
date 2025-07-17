/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/KasumiMercury/mock-todo-server/cmd"
	"github.com/KasumiMercury/mock-todo-server/pid"
	"log"
)

func main() {
	log.Println(pid.PidFile)
	cmd.Execute()
}
