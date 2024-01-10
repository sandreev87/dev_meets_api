package rest

type ErrResponse struct {
	Status string `json:"status" example:"wrong_params, internal_server_error"`
}

type OkResponse struct {
	Status string `json:"status" example:"ok"`
}
