package domain

// Pagination bounds, shared by the service layer (which clamps) and the
// persistence adapter (which refuses to run an unbounded query) so the two
// cannot drift apart.
//
// These mirror catalog-service's domain constants by design: a client should
// get the same default page size whichever API it calls.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)
