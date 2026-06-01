package main

import (
	_ "embed"
)

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}

// func (g *GreetService) OpenBrowser(url string) error {
// 	return application.OpenURL(url)
// }

// func (g *GreetService) CloseWindow() error {
// 	window := application.CurrentWindow()
// 	if window != nil {
// 		return window.Close()
// 	}
// 	return nil
// }
