package model

type ModuleTempl struct {
	URL      string
	Name     string
	Version  string
	Category int
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
