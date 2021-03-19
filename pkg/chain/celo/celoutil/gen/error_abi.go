package gen

// ABI for errors bubbled out from revert calls. Not used directly as errors are
// neither encoded strictly as method calls nor strictly as return values, nor
// strictly as events, but some various bits of it are used for unpacking the
// errors. See ResolveError below.
const errorABIString = "[{\"constant\":true,\"outputs\":[{\"type\":\"string\"}],\"inputs\":[{\"name\":\"message\", \"type\":\"string\"}],\"name\":\"Error\", \"type\": \"function\"}]"
