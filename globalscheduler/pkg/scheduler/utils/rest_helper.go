/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
