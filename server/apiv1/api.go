package apiv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/axllent/mailpit/config"
	"github.com/axllent/mailpit/storage"
	"github.com/axllent/mailpit/utils/logger"
	"github.com/gorilla/mux"
)

// GetMessages returns a paginated list of messages as JSON
func GetMessages(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/messages messages GetMessages
	//
	// # List messages
	//
	// Returns messages from the mailbox ordered from newest to oldest.
	//
	//	Produces:
	//	- application/json
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: start
	//	    in: query
	//	    description: pagination offset
	//	    required: false
	//	    type: integer
	//	    default: 0
	//	  + name: limit
	//	    in: query
	//	    description: limit results
	//	    required: false
	//	    type: integer
	//	    default: 50
	//
	//	Responses:
	//		200: MessagesSummaryResponse
	//		default: ErrorResponse
	start, limit := getStartLimit(r)

	messages, err := storage.List(start, limit)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	stats := storage.StatsGet()

	var res MessagesSummary

	res.Start = start
	res.Messages = messages
	res.Count = len(messages)
	res.Total = stats.Total
	res.Unread = stats.Unread
	res.Tags = stats.Tags

	bytes, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

// Search returns the latest messages as JSON
func Search(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/search messages MessagesSummary
	//
	// # Search messages
	//
	// Returns the latest messages matching a search.
	//
	//	Produces:
	//	- application/json
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: query
	//	    in: query
	//	    description: search query
	//	    required: true
	//	    type: string
	//	  + name: limit
	//	    in: query
	//	    description: limit results
	//	    required: false
	//	    type: integer
	//	    default: 50
	//
	//	Responses:
	//		200: MessagesSummaryResponse
	//		default: ErrorResponse
	search := strings.TrimSpace(r.URL.Query().Get("query"))
	if search == "" {
		httpError(w, "Error: no search query")
		return
	}

	start, limit := getStartLimit(r)

	messages, err := storage.Search(search, start, limit)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	stats := storage.StatsGet()

	var res MessagesSummary

	res.Start = 0
	res.Messages = messages
	res.Count = len(messages)
	res.Total = stats.Total
	res.Unread = stats.Unread
	res.Tags = stats.Tags

	bytes, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

// GetMessage (method: GET) returns the Message as JSON
func GetMessage(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/message/{ID} message Message
	//
	// # Get message summary
	//
	// Returns the summary of a message, marking the message as read.
	//
	//	Produces:
	//	- application/json
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ID
	//	    in: path
	//	    description: message id
	//	    required: true
	//	    type: string
	//
	//	Responses:
	//		200: Message
	//		default: ErrorResponse

	vars := mux.Vars(r)

	id := vars["id"]

	msg, err := storage.GetMessage(id)
	if err != nil {
		fourOFour(w)
		return
	}

	bytes, _ := json.Marshal(msg)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

// DownloadAttachment (method: GET) returns the attachment data
func DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/message/{ID}/part/{PartID} message Attachment
	//
	// # Get message attachment
	//
	// This will return the attachment part using the appropriate Content-Type.
	//
	//	Produces:
	//	- application/*
	//	- image/*
	//	- text/*
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ID
	//	    in: path
	//	    description: message id
	//	    required: true
	//	    type: string
	//	  + name: PartID
	//	    in: path
	//	    description: attachment part id
	//	    required: true
	//	    type: string
	//
	//	Responses:
	//		200: BinaryResponse
	//		default: ErrorResponse

	vars := mux.Vars(r)

	id := vars["id"]
	partID := vars["partID"]

	a, err := storage.GetAttachmentPart(id, partID)
	if err != nil {
		fourOFour(w)
		return
	}
	fileName := a.FileName
	if fileName == "" {
		fileName = a.ContentID
	}

	w.Header().Add("Content-Type", a.ContentType)
	w.Header().Set("Content-Disposition", "filename=\""+fileName+"\"")
	_, _ = w.Write(a.Content)
}

// GetHeaders (method: GET) returns the message headers as JSON
func GetHeaders(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/message/{ID}/headers message Headers
	//
	// # Get message headers
	//
	// Returns the message headers as an array.
	//
	//	Produces:
	//	- application/json
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ID
	//	    in: path
	//	    description: message id
	//	    required: true
	//	    type: string
	//
	//	Responses:
	//	  200: MessageHeaders
	//	  default: ErrorResponse

	vars := mux.Vars(r)

	id := vars["id"]

	data, err := storage.GetMessageRaw(id)
	if err != nil {
		fourOFour(w)
		return
	}

	reader := bytes.NewReader(data)
	m, err := mail.ReadMessage(reader)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	bytes, _ := json.Marshal(m.Header)

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

// DownloadRaw (method: GET) returns the full email source as plain text
func DownloadRaw(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/message/{ID}/raw message Raw
	//
	// # Get message source
	//
	// Returns the full email source as plain text.
	//
	//	Produces:
	//	- text/plain
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ID
	//	    in: path
	//	    description: message id
	//	    required: true
	//	    type: string
	//
	//	Responses:
	//		200: TextResponse
	//		default: ErrorResponse

	vars := mux.Vars(r)

	id := vars["id"]

	dl := r.FormValue("dl")

	data, err := storage.GetMessageRaw(id)
	if err != nil {
		fourOFour(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if dl == "1" {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+id+".eml\"")
	}
	_, _ = w.Write(data)
}

// ReleaseMessage (method: GET) will release a message to the relay server
func ReleaseMessage(w http.ResponseWriter, r *http.Request) {
	// swagger:route GET /api/v1/message/{ID}/release
	//
	// # Get releases message to relay server
	//
	// Releases the message to the designated SMTP Server.
	//
	//	Produces:
	//	- text/plain
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ID
	//	    in: path
	//	    description: message id
	//	    required: true
	//	    type: string
	//
	//	Responses:
	//		200: OKResponse
	//		default: ErrorResponse

	vars := mux.Vars(r)

	id := vars["id"]

	msg, err := storage.GetMessage(id)
	if err != nil {
		fourOFour(w)
		return
	}
	var tos []string
	for _, t := range msg.To {
		tos = append(tos, t.Address)
	}

	err = smtp.SendMail(config.SMTPRelayServer, nil, msg.From.Address, tos, msg.Raw)
	if err != nil {
		logger.Log().Errorf("[smtp] error relaying email (%s) ", err)
	}

	w.Header().Add("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}

// DeleteMessages (method: DELETE) deletes all messages matching IDS.
func DeleteMessages(w http.ResponseWriter, r *http.Request) {
	// swagger:route DELETE /api/v1/messages messages Delete
	//
	// # Delete messages
	//
	// If no IDs are provided then all messages are deleted.
	//
	//	Consumes:
	//	- application/json
	//
	//	Produces:
	//	- text/plain
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ids
	//	    in: body
	//	    description: Message ids to delete
	//	    required: false
	//	    type: DeleteRequest
	//
	//	Responses:
	//		200: OKResponse
	//		default: ErrorResponse

	decoder := json.NewDecoder(r.Body)
	var data struct {
		IDs []string
	}
	err := decoder.Decode(&data)
	if err != nil || len(data.IDs) == 0 {
		if err := storage.DeleteAllMessages(); err != nil {
			httpError(w, err.Error())
			return
		}
	} else {
		for _, id := range data.IDs {
			if err := storage.DeleteOneMessage(id); err != nil {
				httpError(w, err.Error())
				return
			}
		}
	}

	w.Header().Add("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}

// SetReadStatus (method: PUT) will update the status to Read/Unread for all provided IDs
// If no IDs are provided then all messages are updated.
func SetReadStatus(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /api/v1/messages messages SetReadStatus
	//
	// # Set read status
	//
	// If no IDs are provided then all messages are updated.
	//
	//	Consumes:
	//	- application/json
	//
	//	Produces:
	//	- text/plain
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ids
	//	    in: body
	//	    description: Message ids to update
	//	    required: false
	//	    type: SetReadStatusRequest
	//
	//	Responses:
	//		200: OKResponse
	//		default: ErrorResponse

	decoder := json.NewDecoder(r.Body)

	var data struct {
		Read bool
		IDs  []string
	}

	err := decoder.Decode(&data)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	ids := data.IDs

	if len(ids) == 0 {
		if data.Read {
			err := storage.MarkAllRead()
			if err != nil {
				httpError(w, err.Error())
				return
			}
		} else {
			err := storage.MarkAllUnread()
			if err != nil {
				httpError(w, err.Error())
				return
			}
		}
	} else {
		if data.Read {
			for _, id := range ids {
				if err := storage.MarkRead(id); err != nil {
					httpError(w, err.Error())
					return
				}
			}
		} else {
			for _, id := range ids {
				if err := storage.MarkUnread(id); err != nil {
					httpError(w, err.Error())
					return
				}
			}
		}
	}

	w.Header().Add("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}

// SetTags (method: PUT) will set the tags for all provided IDs
func SetTags(w http.ResponseWriter, r *http.Request) {
	// swagger:route PUT /api/v1/tags tags SetTags
	//
	// # Set message tags
	//
	// To remove all tags from a message, pass an empty tags array.
	//
	//	Consumes:
	//	- application/json
	//
	//	Produces:
	//	- text/plain
	//
	//	Schemes: http, https
	//
	//	Parameters:
	//	  + name: ids
	//	    in: body
	//	    description: Message ids to update
	//	    required: true
	//	    type: SetTagsRequest
	//
	//	Responses:
	//		200: OKResponse
	//		default: ErrorResponse

	decoder := json.NewDecoder(r.Body)

	var data struct {
		Tags []string
		IDs  []string
	}

	err := decoder.Decode(&data)
	if err != nil {
		httpError(w, err.Error())
		return
	}

	ids := data.IDs

	if len(ids) > 0 {
		for _, id := range ids {
			if err := storage.SetTags(id, data.Tags); err != nil {
				httpError(w, err.Error())
				return
			}
		}
	}

	w.Header().Add("Content-Type", "text/plain")
	_, _ = w.Write([]byte("ok"))
}

// FourOFour returns a basic 404 message
func fourOFour(w http.ResponseWriter) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "404 page not found")
}

// HTTPError returns a basic error message (400 response)
func httpError(w http.ResponseWriter, msg string) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, msg)
}

// Get the start and limit based on query params. Defaults to 0, 50
func getStartLimit(req *http.Request) (start int, limit int) {
	start = 0
	limit = 50

	s := req.URL.Query().Get("start")
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		start = n
	}

	l := req.URL.Query().Get("limit")
	if n, err := strconv.Atoi(l); err == nil && n > 0 {
		limit = n
	}

	return start, limit
}
