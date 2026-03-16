package dto

type ErrorResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Fields  []field `json:"fields"`
}

type field struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}
