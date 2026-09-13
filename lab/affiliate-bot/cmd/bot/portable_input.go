package main

// M03/M04 fixture imports deliberately exercise their own 1 MiB semantic
// store limit (including an over-limit STORE_ERROR). Keep the transport guard
// above that contract while still bounding every generic CLI read; selected
// report imports use this same 16 MiB ceiling explicitly.
const maxGeneralPortableInputBytes int64 = 16 << 20

// readGeneralPortableInput is the common boundary for caller-supplied files
// outside M07 and the M08-M11 mission commands. A valid JSON/CSV payload does
// not become trusted merely because a CLI later joins it to canonical history,
// actions, or outcomes; reject a post-open pathname swap before decoding or
// persistence.
func readGeneralPortableInput(path string) ([]byte, error) {
	b, _, err := readStableRegularFileLimit(path, maxGeneralPortableInputBytes)
	return b, err
}
