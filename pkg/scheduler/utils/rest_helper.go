package utils

import (
	"fmt"

	"github.com/emicklei/go-restful"
)

// ErrorResponse defines http error response body
type ErrorResponse struct {
	ErrCode    string `json:"error_code"`
	ErrMessage string `json:"error_message"`
}

// WriteFailedJSONResponse returns error response
func WriteFailedJSONResponse(resp *restful.Response, httpStatusCode int, apiError APIError) {
	err := ErrorResponse{
		ErrCode:    fmt.Sprintf("%s.0%d%04d", apiError.ErrorPrefix, apiError.ErrorType, apiError.ErrorCode),
		ErrMessage: apiError.ErrorMsg,
	}

	resp.WriteHeaderAndEntity(httpStatusCode, err)
}

// WriteFailedJSONResponseByMsg returns error response
func WriteFailedJSONResponseByMsg(resp *restful.Response, httpStatusCode int, msg string) {
	err := ErrorResponse{
		ErrCode:    fmt.Sprintf("%d", httpStatusCode),
		ErrMessage: msg,
	}

	resp.WriteHeaderAndEntity(httpStatusCode, err)
}

// WriteSuccessResponse returns normal response
func WriteSuccessResponse(resp *restful.Response, httpStatusCode int, message string) {
	resp.WriteHeader(httpStatusCode)
	resp.Write([]byte(message))
}
