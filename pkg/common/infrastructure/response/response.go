package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
}

type RedirectResponse struct {
	RedirectURL string `json:"redirect_url"`
}

func WriteErrorResponse(err error, w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: err.Error()})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write(data)
}

func WriteSuccessWithRedirectResponse(url string, w http.ResponseWriter) {
	data, err := json.Marshal(RedirectResponse{RedirectURL: url})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func WriteSuccessResponse(w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: http.StatusText(http.StatusOK)})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func WriteForbiddenResponse(w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: http.StatusText(http.StatusForbidden)})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write(data)
}

func WriteNotFoundResponse(err error, w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: err.Error()})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(data)
}

func WriteDuplicateRequestResponse(err error, w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: err.Error()})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	_, _ = w.Write(data)
}

func WriteBadRequestResponse(err error, w http.ResponseWriter) {
	data, err := json.Marshal(Response{Message: err.Error()})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write(data)
}

func WriteUnauthorizedResponse(message string, w http.ResponseWriter) {
	if message == "" {
		message = "login required"
	}
	data, err := json.Marshal(Response{Message: message})
	if err != nil {
		WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(data)
}
