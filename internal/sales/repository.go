package sales

// No direct DB access needed here: bill creation/listing goes through
// billing.Service and product listing through product.Service (see
// handler.go and cmd/api/main.go for wiring).
