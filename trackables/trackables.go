package trackables

// Root ....
type Root struct{}

// GetProfileName ...
func (r Root) GetProfileName() string {
	return "RootProfile"
}

// GetTagArray ...
func (r Root) GetTagArray() []string {
	return []string{"tag1", "tag2"}
}
