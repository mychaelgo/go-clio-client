package datamodels

type Client struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Type      string `json:"type"`
	Initials  string `json:"initials"`
}
