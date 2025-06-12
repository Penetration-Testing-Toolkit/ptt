package model

import "github.com/Penetration-Testing-Toolkit/ptt/shared"

type ModuleTempl struct {
	ID       string
	URL      string
	Name     string
	Version  string
	Category int
	Metadata []*shared.Metadata
}

var Categories = []string{
	"Misc",
	"Scanner",
	"Password",
	"Shell",
	"Exploit",
	"Web",
	"Social Engineering",
	"Forensic",
	"Reporting",
}
