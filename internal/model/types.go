package model

type Dependency struct {
	Path           string `json:"path"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Indirect       bool   `json:"indirect"`
}

type Result struct {
	Module    string       `json:"module"`
	GoVersion string       `json:"go_version"`
	Updates   []Dependency `json:"updates"`
}

type GoListModule struct {
	Path     string        `json:"Path"`
	Version  string        `json:"Version"`
	Main     bool          `json:"Main"`
	Indirect bool          `json:"Indirect"`
	Update   *ModuleUpdate `json:"Update"`
}

type ModuleUpdate struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
}
