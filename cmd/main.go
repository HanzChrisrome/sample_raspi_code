package main

import (
	"fmt"

	"github.com/otis-co-ltd/aihub-recorder/internal/pi"
	"github.com/otis-co-ltd/aihub-recorder/internal/wsclient"
)

func main() {
	piID := pi.GetPiId()
	fmt.Println("Pi ID:", piID)

	wsclient.Start(piID)
}
