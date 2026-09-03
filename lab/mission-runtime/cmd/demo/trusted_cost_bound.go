package main

// TrustedCostBound is the stage-neutral runtime name for the trusted cost artifact
// used by both governed canary (M10) and governed production (M11).
// CanaryCostBound remains the underlying Go representation inside m10.go so this
// semantic cleanup does not create a second hash implementation.
type TrustedCostBound = CanaryCostBound

func ComputeTrustedCostBoundHash(c TrustedCostBound) string {
	return ComputeCanaryCostBoundHash(c)
}

func SealTrustedCostBound(c TrustedCostBound) TrustedCostBound {
	return SealCanaryCostBound(c)
}
