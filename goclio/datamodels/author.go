package datamodels

type Author struct {
	Id   int    `json:"id"`
	Etag string `json:"etag"`
	Name string `json:"name"`
}
