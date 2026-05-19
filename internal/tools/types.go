package tools

import "time"

type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone"`
	Address   *string   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	// Region tags the customer's data residency / authz partition. A value of
	// "atlantis" is gated by the customers:atlantis:read permission; the
	// row-level filter in search_customer.go and lookup_customer.go excludes
	// these rows from callers who lack that perm.
	Region string `json:"region"`
}

type Order struct {
	ID          int64     `json:"id"`
	CustomerID  int64     `json:"customer_id"`
	Status      string    `json:"status"`
	TotalCents  int64     `json:"total_cents"`
	Currency    string    `json:"currency"`
	PlacedAt    time.Time `json:"placed_at"`
}

type Transaction struct {
	ID          int64     `json:"id"`
	CustomerID  int64     `json:"customer_id"`
	AmountCents int64     `json:"amount_cents"`
	Kind        string    `json:"kind"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderItem struct {
	ID        int64  `json:"id"`
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	UnitCents int64  `json:"unit_cents"`
}

type OrderWithItems struct {
	Order
	LineItems []OrderItem `json:"line_items"`
}
