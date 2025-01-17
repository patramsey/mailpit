package apiv1

// These structs are for the purpose of defining swagger HTTP responses

// Application information
// swagger:response InfoResponse
type infoResponse struct {
	// Application information
	Body appInformation
}

// Message summary
// swagger:response MessagesSummaryResponse
type messagesSummaryResponse struct {
	// The message summary
	// in: body
	Body MessagesSummary
}

// Message headers
// swagger:model MessageHeaders
type messageHeaders map[string][]string

// Delete request
// swagger:model DeleteRequest
type deleteRequest struct {
	// ids
	// in:body
	IDs []string `json:"ids"`
}

// Set read status request
// swagger:model SetReadStatusRequest
type setReadStatusRequest struct {
	// Read status
	Read bool `json:"read"`

	// ids
	// in:body
	IDs []string `json:"ids"`
}

// Set tags request
// swagger:model SetTagsRequest
type setTagsRequest struct {
	// Tags
	// in:body
	Tags []string `json:"tags"`

	// ids
	// in:body
	IDs []string `json:"ids"`
}

// Binary data reponse inherits the attachment's content type
// swagger:response BinaryResponse
type binaryResponse struct {
	// in: body
	Body string
}

// Plain text response
// swagger:response TextResponse
type textResponse struct {
	// in: body
	Body string
}

// Error reponse
// swagger:response ErrorResponse
type errorResponse struct {
	// The error message
	// in: body
	Body string
}

// Plain text "ok" reponse
// swagger:response OKResponse
type okResponse struct {
	// Default reponse
	// in: body
	Body string
}
