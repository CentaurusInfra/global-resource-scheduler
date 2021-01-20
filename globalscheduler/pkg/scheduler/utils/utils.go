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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"testing"
	"unicode/utf8"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

const (
	// UUIDCheckPattern uses to check server name
	UUIDCheckPattern = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
	// DefaultPriorityWhenNoDefaultClassExists is used to set priority of pods
	// that do not specify any priority class and there is no priority class
	// marked as default.
	DefaultPriorityWhenNoDefaultClassExists = 0
)

// AssertIntEqual triggers testing error if the expect and actual int value are not the same.
func AssertIntEqual(t *testing.T, expect, actual, errMsg string) {
	if expect != actual {
		t.Errorf("%s, expect:%s, actual:%s", errMsg, expect, actual)
	}
}

// GetConfigDirectory Get the configuration file path
func GetConfigDirectory() string {
	configBase := os.Getenv("CONFIG_BASE")
	if configBase == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			configBase = "."
		} else {
			configBase = filepath.Join(dir, "..", "conf")
		}
	}

	return configBase
}

// GetStrFromCtx get value in context by key
func GetStrFromCtx(ctx context.Context, key ContentKey) string {
	if ctx == nil {
		return ""
	}

	val := ctx.Value(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

//GetBoolFromCtx get value of bool from ctx
func GetBoolFromCtx(ctx context.Context, key ContentKey) bool {
	if ctx == nil {
		return false
	}
	val := ctx.Value(key)
	if ret, ok := val.(bool); ok {
		return ret
	}
	return false
}

// NeedAdminFromCtx check if request context need admin client
func NeedAdminFromCtx(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	val := ctx.Value(constants.ContextNeedAdmin)
	if needAdmin, ok := val.(bool); ok {
		return needAdmin
	}
	return false
}

func CatchException() {
	err := recover()
	GetExceptionMsg(err)
}

func GetExceptionMsg(err interface{}) string {
	if err != nil {
		errMsg := fmt.Sprintf("Exception: %+v, call stack: %v", err, string(debug.Stack()))
		fmt.Println(errMsg)
		return errMsg
	}
	return ""
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func InsertFailReason(ctx context.Context, ecID string, failReason error) error {
	return nil
}

// merge map
func MergeMap(args ...map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for _, arg := range args {
		for k, v := range arg {
			newMap[k] = v
		}
	}

	return newMap
}

func GetZoneKey(site *types.Site) string {

	region := site.Region
	zone := site.AvailabilityZone

	if region == "" && zone == "" {
		return ""
	}

	// We include the null character just in case region or failureDomain has a colon
	// (We do assume there's no null characters in a region or failureDomain)
	// As a nice side-benefit, the null character is not printed by fmt.Print or glog
	return region + ":\x00:" + zone
}

func IsContain(allStr []string, dst string) bool {
	for _, one := range allStr {
		if one == dst {
			return true
		}
	}

	return false
}

func StringIsLegal(str string) bool {
	regx, err := regexp.Compile("^[0-9]{1,64}$")
	if err != nil {
		return false
	}
	res := regx.MatchString(str)
	return res
}

func QueryStrIsLegal(uuid string) bool {
	regx, err := regexp.Compile("^[a-zA-Z0-9-'_]{1,64}$")
	if err != nil {
		return false
	}
	res := regx.MatchString(uuid)
	return res
}

func UUIDIsLegal(uuid string) bool {
	regx, err := regexp.Compile("^[a-z0-9]{8}(-[a-z0-9]{4}){3}-[a-z0-9]{12}$")
	if err != nil {
		return false
	}
	res := regx.MatchString(uuid)
	return res
}

func GetJSONString(content interface{}) string {
	jsonStr, err := json.Marshal(content)
	if err != nil {
		return ""
	}
	return string(jsonStr)
}
func isLengthInRange(title string, minLen int, maxLen int) bool {
	length := utf8.RuneCountInString(title)
	return length < minLen || length > maxLen
}

func CheckTitleLengthAndRegexValid(title string, minLen int, maxLen int, regex string) error {
	if strings.EqualFold(title, "") || strings.EqualFold(regex, "") {
		return fmt.Errorf("invalid input.title or regex can't be empty")
	}

	// length check
	if isLengthInRange(title, minLen, maxLen) {
		return fmt.Errorf("title can not match length")
	}

	// regex check
	if match, err := regexp.MatchString(regex, title); err != nil {
		return fmt.Errorf("meet error when title (%s) match regex (%s): %s", title, regex, err.Error())
	} else if !match {
		return fmt.Errorf("title can not match regexp")
	}

	return nil
}

func IsUUIDValid(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("uuid is invalid!It can not be empty")
	}

	if match, err := regexp.MatchString(UUIDCheckPattern, uuid); err != nil {
		return fmt.Errorf("uuid (%s) match regex failed, error: %s", uuid, err.Error())
	} else if !match {
		return fmt.Errorf("uuid (%s) is invalid", uuid)
	}
	return nil
}

// LessFunc is a function that receives two items and returns true if the first
// item should be placed before the second one when the list is sorted.
type LessFunc func(item1, item2 interface{}) bool

// GetStackPriority returns priority of the given stack.
func GetStackPriority(stack *types.Stack) int32 {
	// When priority of a running pod is nil, it means it was created at a time
	// that there was no global default priority class and the priority class
	// name of the pod was empty. So, we resolve to the static default priority.
	return DefaultPriorityWhenNoDefaultClassExists
}

// GetPodFullName returns a name that uniquely identifies a stack.
func GetStackFullName(stack *types.Stack) string {
	// Use underscore as the delimiter because it is not allowed in stack name
	// (DNS subdomain format).
	return stack.Name + "_" + stack.Namespace
}
