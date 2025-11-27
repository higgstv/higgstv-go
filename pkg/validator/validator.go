package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init 初始化驗證器
func Init() {
	validate = validator.New()
	
	// 註冊自訂驗證器
	if err := RegisterCustomValidators(validate); err != nil {
		panic(fmt.Sprintf("Failed to register custom validators: %v", err))
	}
}

// Validate 驗證結構體
func Validate(s interface{}) error {
	if validate == nil {
		Init()
	}
	return validate.Struct(s)
}

// ValidateVar 驗證單一變數
func ValidateVar(field interface{}, tag string) error {
	if validate == nil {
		Init()
	}
	return validate.Var(field, tag)
}

