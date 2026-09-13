package main

const maxGeneralPortableInputBytes int64 = 1 << 20

// readGeneralPortableInput is the common boundary for caller-supplied files
// outside M07 and the M08-M11 mission commands. A valid JSON/CSV payload does
// not become trusted merely because a CLI later joins it to canonical history,
// actions, or outcomes; reject a post-open pathname swap before decoding or
// persistence.
func readGeneralPortableInput(path string) ([]byte, error) {
	b, _, err := readStableRegularFileLimit(path, maxGeneralPortableInputBytes)
	return b, err
}
