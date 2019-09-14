package monitoring

type Alert string
type Warning string

type Module interface {
	IsEnabled() bool
	Run() error
	GetName() string
	GetExecutedCommand() string
	GetAlerts() []Alert
	GetWarnings() []Warning
	GetMessage() string
	GetMeasurements() map[string]interface{}
}
