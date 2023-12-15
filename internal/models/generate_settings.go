package models

type GenerateSettingsRespBody struct {
	Id       string `json:"id"`
	Versions []struct {
		Version   string   `json:"version"`
		Protocols []string `json:"protocols"`
		Platforms []struct {
			Os   string `json:"os"`
			Arch string `json:"arch"`
		} `json:"platforms"`
	} `json:"versions"`
	Warnings interface{} `json:"warnings"`
}

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
