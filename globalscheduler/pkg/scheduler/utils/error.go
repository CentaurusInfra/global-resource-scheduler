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
)

// constants for error type
const (
	ErrGeneral int = 0

	ErrService int = 5
)

//constants for prefix type
const (
	SchedulerPrefix = "GS"
)

// APIError holds API errors for function manager
type APIError struct {
	ErrorPrefix string
	ErrorCode   int
	ErrorType   int
	ErrorMsg    string
}

// GetQuotaError error
func GetQuotaError() APIError {
	return APIError{SchedulerPrefix, 0, ErrGeneral, "get the quota failed!"}
}

// InternalServerError error
func InternalServerError() APIError {
	return APIError{SchedulerPrefix, 1, ErrGeneral, "Internal Server Error!"}
}

// InternalServerWithError error with error msg
func InternalServerWithError(errMsg string) APIError {
	return APIError{SchedulerPrefix, 1, ErrGeneral,
		fmt.Sprintf("Internal Server Error: %s", errMsg)}
}

// UserHasNoAccessError Role without permission
func UserHasNoAccessError() APIError {
	return APIError{SchedulerPrefix, 5, ErrGeneral,
		"the current user have no access to service"}
}

// RequestQueryParamInvalid check request URI error
func RequestQueryParamInvalid() APIError {
	return APIError{SchedulerPrefix, 6, ErrGeneral,
		fmt.Sprintf("the query parameter is illegal")}
}

// RequestBodyParamInvalid check request body error
func RequestBodyParamInvalid(errMsg string) APIError {
	return APIError{SchedulerPrefix, 7, ErrGeneral,
		fmt.Sprintf("The request body parameter is illegal: %s.", errMsg)}
}

// RequestURIInvalidError error
func RequestURIInvalidError() APIError {
	return APIError{SchedulerPrefix, 8, ErrGeneral, "request uri is illegal!"}
}

// DomainIDNotFoundError error
func DomainIDNotFoundError() APIError {
	return APIError{SchedulerPrefix, 9, ErrGeneral,
		"domain_id is not found in user token!"}
}

// RequestBodyInvalid error
func RequestBodyInvalid() APIError {
	return APIError{SchedulerPrefix, 10, ErrGeneral, "the request body is illegal!"}
}

// RequestBodyLackParamError error
func RequestBodyLackParamError() APIError {
	return APIError{SchedulerPrefix, 11, ErrGeneral,
		"the request body lack required parameter!"}
}

// RequestURIParamInvalidError error
func RequestURIParamInvalidError(pathParam string) APIError {
	return APIError{SchedulerPrefix, 30, ErrGeneral,
		fmt.Sprintf("request uri path param (%s) is illegal!Please correct it.", pathParam)}
}

// RequestURIParamEmptyError error
func RequestURIParamEmptyError(pathParamName string) APIError {
	return APIError{SchedulerPrefix, 31, ErrGeneral,
		fmt.Sprintf("request uri path param (%s) can't be empty!Please input it.", pathParamName)}
}

// AllocationInvalidBody body
func AllocationInvalidBody() APIError {
	return APIError{SchedulerPrefix, 1, ErrService, "Invalid body"}
}

// SchedulerInitFailed scheduler failed
func SchedulerInitFailed() APIError {
	return APIError{SchedulerPrefix, 2, ErrService, "Scheduler init failed"}
}

// ServiceNotFound scheduler failed
func ServiceNotFound() APIError {
	return APIError{SchedulerPrefix, 3, ErrService, "Service not found"}
}

// DelayCoverCitiesFailed scheduler failed
func DelayCoverCitiesFailed() APIError {
	return APIError{SchedulerPrefix, 4, ErrService,
		"get delay cover cities data failed"}
}

func NetworkQualityListFailed() APIError {
	return APIError{SchedulerPrefix, 5, ErrService,
		"get network quality list data failed"}
}
