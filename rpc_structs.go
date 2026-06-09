package main


type StartServerRequest struct {
	ServerName string
	MinecraftVersion string
	ServerSoftware string
	SoftwareVersion string
}

type StartServerResponse struct {
	Status string
	IPAddress string
	ID string
	Port int
}

type PingResponse struct {
	Status string
	Name string
}

type CompletionNotice struct {
	Status string
	Complete bool
}
