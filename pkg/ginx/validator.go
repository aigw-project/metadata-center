// Copyright The AIGW Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package ginx

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	mutuallyExclusiveTag = "mutually_exclusive"
	eitherOrTag          = "either_or"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation(eitherOrTag, validateEitherOrTag)
		_ = v.RegisterValidation(mutuallyExclusiveTag, validateMutuallyExclusiveTag)
	}
}

func validateEitherOrTag(fl validator.FieldLevel) bool {
	otherField := fl.Param()
	structValue := fl.Parent()

	currentValue := fl.Field()
	otherValue := structValue.FieldByName(otherField)

	hasCurrent := !currentValue.IsZero() && currentValue.Len() > 0
	hasOther := !otherValue.IsZero() && otherValue.Len() > 0

	// Mutually exclusive logic: cannot exist simultaneously
	return !(hasCurrent && hasOther)
}

func validateMutuallyExclusiveTag(fl validator.FieldLevel) bool {
	otherField := fl.Param()
	structValue := fl.Parent()

	currentValue := fl.Field()
	otherValue := structValue.FieldByName(otherField)

	hasCurrent := !currentValue.IsZero() && currentValue.Len() > 0
	hasOther := !otherValue.IsZero() && otherValue.Len() > 0

	// Must have exactly one, cannot have both or none
	return hasCurrent != hasOther // Exactly one must be true
}
