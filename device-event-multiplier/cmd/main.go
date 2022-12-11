// Sources for https://watermill.io/docs/getting-started/
package main

import "device-event-multiplier/internal/service"

// main - entry point.
func main() {
	service.New().Start()
}
