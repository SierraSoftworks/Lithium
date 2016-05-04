package license

// Product tracks a product which makes use of Lithium for licensing
// purposes.
type Product struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
}
