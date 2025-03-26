package domains

type ValidatorGrpcPort interface {
	Valid() bool
	AddError(key string, msg string)
	Validate(condition bool, key string, msg string)
	ValidatorErrors() map[string]string
}
