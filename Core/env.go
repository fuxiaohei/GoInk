package Core

const (
	Dev  string = "development"
	Prod string = "production"
)

var Env = Dev

func IsDev() bool {
	return Env == Dev
}

func IsProd() bool {
	return Env == Prod
}
