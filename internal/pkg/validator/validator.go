package validator

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

const (
	InvalidType  = "invalid type"
	InvalidEmail = "invalid email"
	// maxUploadSizeBytes is used by image file validation (20 MB).
	maxUploadSizeBytes = 20 << 20
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string
	Rules  map[string][]ValidationRule
}

type ValidationRule struct {
	Field string
	Rules []func(any) (bool, string)
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
		Rules:  make(map[string][]ValidationRule),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// For future checks.
func In(value string, list ...string) bool {
	return slices.Contains(list, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

func ValidateStruct(v *Validator, data any, rules []ValidationRule) {
	val := reflect.ValueOf(data).Elem()

	for _, rule := range rules {
		field := val.FieldByName(rule.Field)
		if !field.IsValid() {
			continue
		}

		for _, validationFunc := range rule.Rules {
			ok, message := validationFunc(field.Interface())
			v.Check(ok, rule.Field, message)
		}
	}
}

//lint:ignore U1000 pre-existing dead code, do not delete
func validCategory(value any) (bool, string) {
	if value == nil {
		return false, "must choose at least one category"
	}

	rv := reflect.ValueOf(value)
	// unwrap pointers and interfaces
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return false, "must choose at least one category"
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len() > 0, "must choose at least one category"
	default:
		return false, InvalidType
	}
}

func validateImageFile(value any) (bool, string) {
	// Allowed content types mirror client-side checks
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	if value == nil {
		// no file provided -> optional
		return true, ""
	}

	rv := reflect.ValueOf(value)
	// // unwrap pointers and interfaces safely
	// for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
	// 	if rv.IsNil() {
	// 		return true, ""
	// 	}
	// 	rv = rv.Elem()
	// }

	if rv.Kind() != reflect.Struct {
		return false, InvalidType
	}

	// Expect a struct with a Header field of type *multipart.FileHeader
	hf := rv.FieldByName("Header")
	if !hf.IsValid() || hf.IsZero() {
		// no header -> treat as no file provided
		return true, ""
	}

	hdr, ok := hf.Interface().(*multipart.FileHeader)
	if !ok {
		return false, "invalid file header type"
	}

	// Check content type
	contentType := hdr.Header.Get("Content-Type")
	if !allowed[contentType] {
		return false, "invalid file type; only JPEG, PNG and GIF are allowed"
	}

	// Check size
	if hdr.Size > maxUploadSizeBytes {
		return false, "file too large; maximum size is 20MB"
	}

	return true, ""
}

func required(value any) (bool, string) {
	switch v := value.(type) {
	case string:
		if v != "" {
			return true, ""
		}
		return false, "must be provided"
	case int:
		return true, ""
	default:
		return false, InvalidType
	}
}

func optional(validationFunc func(any) (bool, string)) func(any) (bool, string) {
	return func(value any) (bool, string) {
		str, ok := value.(string)
		if !ok {
			return false, "field must be a string"
		}

		if str == "" {
			return true, ""
		}

		return validationFunc(value)
	}
}

func minLength(minimumLenght int) func(any) (bool, string) {
	return func(value any) (bool, string) {
		str, ok := value.(string)
		if !ok {
			return false, InvalidType
		}
		if len(str) >= minimumLenght {
			return true, ""
		}
		return false, fmt.Sprintf("must be at least %d characters long", minimumLenght)
	}
}

func maxLength(maximumLenght int) func(any) (bool, string) {
	return func(value any) (bool, string) {
		str, ok := value.(string)
		if !ok {
			return false, InvalidType
		}
		if len(str) <= maximumLenght {
			return true, ""
		}
		return false, fmt.Sprintf("must be %d characters maximum", maximumLenght)
	}
}

func isPositiveInt(value any) (bool, string) {
	num, ok := value.(int)
	if !ok {
		return false, InvalidType
	}
	if num > 0 {
		return true, ""
	}
	return false, "must be a positive integer"
}

func maxInt(limit int) func(any) (bool, string) {
	return func(value any) (bool, string) {
		num, ok := value.(int)
		if !ok {
			return false, InvalidType
		}
		if num <= limit {
			return true, ""
		}
		return false, fmt.Sprintf("must be less than or equal to %d", limit)
	}
}

func hasLower(value any) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	for _, c := range str {
		if unicode.IsLower(c) {
			return true, ""
		}
	}
	return false, "must contain at least one lowercase letter"
}

func hasUpper(value any) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	for _, c := range str {
		if unicode.IsUpper(c) {
			return true, ""
		}
	}
	return false, "must contain at least one uppercase letter"
}

func hasDigit(value any) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	for _, c := range str {
		if unicode.IsDigit(c) {
			return true, ""
		}
	}
	return false, "must contain at least one digit"
}

func hasSpecial(value any) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	for _, c := range str {
		if !unicode.IsLower(c) && !unicode.IsUpper(c) && !unicode.IsDigit(c) {
			return true, ""
		}
	}
	return false, "must contain at least one special character"
}

func validEmail(value any) (bool, string) {
	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	if Matches(str, EmailRX) {
		return true, ""
	}
	return false, InvalidEmail
}

func (v *Validator) ToStringErrors() string {
	strError := ""
	var strErrorSb310 strings.Builder
	for key, value := range v.Errors {
		strErrorSb310.WriteString(key + ": " + value + " ")
	}
	strError += strErrorSb310.String()
	return strings.TrimSpace(strError)
}

func validImagePath(value any) (bool, string) {
	validImageExtensions := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
	}

	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	ext := strings.ToLower(filepath.Ext(str))
	if validImageExtensions[ext] {
		return true, ""
	}
	return false, "must be a valid image file"
}

// var validCategories = map[string]bool{
// 	"General Discussion": true,
// 	"Feedback":           true,
// 	"Off-Topic":          true,
// }

// func validCategory(value any) (bool, string) {
// 	str, ok := value.(string)
// 	if !ok {
// 		return false, InvalidType
// 	}
// 	return validCategories[str], "must be a valid category"
// }

// ValidateCreateComment validates the create comment request

func validTopicOrderBy(value any) (bool, string) {
	orderByWhitelist := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"vote_score": true,
	}

	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	if str == "" {
		return true, ""
	}
	if orderByWhitelist[str] {
		return true, ""
	}
	return false, "must be a valid order by field"
}

func validCategoryOrderBy(value any) (bool, string) {
	orderByWhitelist := map[string]bool{
		"name":       true,
		"created_by": true,
		"created_at": true,
	}

	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	if str == "" {
		return true, ""
	}
	if orderByWhitelist[str] {
		return true, ""
	}
	return false, "must be a valid order by field"
}

func validOrder(value any) (bool, string) {
	orderWhiteLIST := map[string]bool{
		"DESC": true,
		"ASC":  true,
	}

	str, ok := value.(string)
	if !ok {
		return false, InvalidType
	}
	if str == "" {
		return true, ""
	}
	upper := strings.ToUpper(str)
	if orderWhiteLIST[upper] {
		return true, ""
	}
	return false, "must be a valid order field"
}
