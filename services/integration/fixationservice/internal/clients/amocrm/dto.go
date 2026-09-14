package amocrm

type fixationPayload struct { //надо нормальный мок а не как щас, не надо чтоб в апи одно а мок другой сильно
	FixationID string `json:"fixation_id"`
	AgencyID   string `json:"agency_id"`
	ExpiresAt  string `json:"expires_at"`
}
type leadRequest struct {
	Name       string `json:"name"`
	Price      int    `json:"price"`
	PipelineID int    `json:"pipeline_id"`
	StatusID   int    `json:"status_id"`
}
type leadsResponse struct {
	Embedded struct {
		Leads []struct {
			ID int `json:"id"`
		} `json:"leads"`
	} `json:"_embedded"`
}
