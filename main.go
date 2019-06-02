package main

import (
    app "news-prettifier-go-server/app"
)
func main() {
	a := app.App{}
	a.Initialize()
	a.Run(":8000")
}