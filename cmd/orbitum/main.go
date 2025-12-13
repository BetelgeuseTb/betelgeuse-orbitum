package main

import (
	"fmt"
	reader "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils"
)

const (
	LOGO_PATH = "internal/resources/logo.txt"
)

func main() {
	data, _ := reader.NewFileReader().ReadFile(LOGO_PATH)
	fmt.Println(string(data))
	fmt.Println("Starting application ...")
	fmt.Println("Application finished.")
}
