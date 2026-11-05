package checkly

// Location is a public location that checks can run from: Region is its
// code, such as "us-east-1", and Name a display name, such as "N. Virginia".
// No [Client] method returns it.
type Location struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}
