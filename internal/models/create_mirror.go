package models

type Version struct {
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
	Platforms []struct {
		OS   string `json:"os"`
		Arch string `json:"arch"`
	} `json:"platforms"`
}

type Module struct {
	ID       string    `json:"id"`
	Versions []Version `json:"versions"`
}

// ErrorVersions If version downloaded after first check - provider will be "ok" value, else no changes(in tmp.json file)
type ErrorVersions struct {
	Provider string `json:"provider"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
}
