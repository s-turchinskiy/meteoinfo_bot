package service

type Servicer interface {
	GetWeather(city string) (result string, err error)
}
