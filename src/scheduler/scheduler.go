package scheduler

type scheduler struct {
	trigger  string     `json:"trigger"`
	commands []commands `json:"commands"`
}

type commands struct {
	url   string `json:"url"`
	hubid string `json:"hubid"`
}

func New(trigger string, commands []commands) scheduler {
	t := scheduler{trigger, commands}
	return t
}
