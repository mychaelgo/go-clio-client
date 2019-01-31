package datamodels

type CustomFieldValues struct {
	Id                string      `json:"id"`
	CustomField       CustomField `json:"custom_field"`
	FieldName         string      `json:"field_name"`
	Value             string      `json:"value"`
	FieldType         string      `json:"field_type"`
	FieldRequired     bool        `json:"field_required"`
	FieldDisplayed    bool        `json:"field_displayed"`
	FieldDisplayOrder int         `json:"field_display_order"`
}
