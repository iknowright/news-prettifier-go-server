package main

import (
    app "./app"
)
func main() {
	a := app.App{}
	a.Initialize()
	a.Run(":8000")
}