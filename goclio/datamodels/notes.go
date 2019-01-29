package datamodels

type Notes struct {
	Id      int     `json:"id"`
	Date    string  `json:"date"`
	Subject string  `json:"subject"`
	Detail  string  `json:"detail"`
	Author  *Author `json:"author"`
	Matter  *Matter `json:"matter"`
}
