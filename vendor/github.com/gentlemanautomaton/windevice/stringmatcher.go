package windevice

// StringMatcher is an interface capable of matching strings.
type StringMatcher interface {
	Match(string) bool
}
