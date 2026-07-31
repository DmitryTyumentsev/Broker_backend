package dto

type ErrorResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Fields  []Field `json:"fields,omitempty"`
}

type Field struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
