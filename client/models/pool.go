package models

type Pool struct {
    Name        string      `json:"name"`
    Id          *string     `json:"id"`
	Size        int         `json:"maxAgents"`
}
