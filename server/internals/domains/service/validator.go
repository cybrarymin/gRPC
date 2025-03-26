package domains

type Validator struct {
	Errors map[string]string
}

func NewValidator() Validator {
	return Validator{
		Errors: make(map[string]string),
	}
}

// checks if the validator has any error or not. if not return true which mean datas are valid
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error with its key to the validator map of errors
func (v *Validator) AddError(key string, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

// validate based on the condition. if condition doesn't meet it adds an error to list of validator errors
func (v *Validator) Validate(condition bool, key string, msg string) {
	if !condition {
		v.AddError(key, msg)
	}
}

// return all the errors
func (v *Validator) ValidatorErrors() map[string]string {
	return v.Errors
}
