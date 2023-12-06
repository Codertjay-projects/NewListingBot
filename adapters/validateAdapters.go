package adapters

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"regexp"
	"strings"
)

type ValidateInterface interface {
	ValidateData(model interface{}) map[string]string
	IsEmail(field, email string) error
}

type ValidateStruct struct {
}

func NewValidate() ValidateInterface {
	return &ValidateStruct{}
}

func (v *ValidateStruct) ValidateData(model interface{}) map[string]string {
	/* this is used to validate models in which  we used instead of the creation serializer*/
	// Validate the login request struct
	var validate = validator.New()

	if err := validate.Struct(model); err != nil {
		// Cast the error to validator.ValidationErrors to access the actual errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)

			for _, err := range validationErrors {
				// Add each error message to the errorMessages map
				errorMessage := strings.Split(err.Error(), err.Field()+"'")
				errorMessages[err.Field()] = errorMessage[2]
			}

			// Return the error messages map
			return errorMessages
		}
	}

	return nil
}

func (v *ValidateStruct) IsEmail(field, email string) error {

	var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if !emailRegexp.MatchString(email) {
		return errors.New("not a valid email")
	}

	return nil
}
