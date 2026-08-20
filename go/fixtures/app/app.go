package app

// Greeting gives the toolchain's checks something to compile, vet, and test.
// An empty package satisfies all three without exercising any of them.
func Greeting() string {
	return "hello"
}
