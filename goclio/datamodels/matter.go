package datamodels

type Matter struct {
	Id                int                 `json:"id"`
	DisplayNumber     string              `json:"display_number"`
	Description       string              `json:"description"`
	Status            string              `json:"status"`
	Location          string              `json:"location"`
	ClientReference   string              `json:"client_reference"`
	Client            *Client             `json:"client"`
	PracticeArea      *PracticeArea       `json:"practice_area"`
	Group             *Group              `json:"group"`
	CustomFieldValues []CustomFieldValues `json:"custom_field_values"`
}
