package dskvs

func init() {
	// Initialize maps according to dskvs_settings.json

	// catch SIGUSR1, reload dskvs_settings.json when caught

	// when load of maps is complete, start janitor goroutine
}
