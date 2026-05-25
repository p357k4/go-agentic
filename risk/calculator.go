package risk

// Calculator implements the mathematical anomaly risk assessment.
// It uses a deterministic multi-factor formula:
// Arisk = phi * Vgeo + psi * Dval + omega * Ibeh
// where:
// - phi = 0.4 (weight of geographical anomaly)
// - psi = 0.4 (weight of transaction value anomaly)
// - omega = 0.2 (weight of behavioral anomaly)
type Calculator struct {
	phi   float64
	psi   float64
	omega float64
}

// NewCalculator initializes the risk calculator with standard parameters.
func NewCalculator() *Calculator {
	return &Calculator{
		phi:   0.4,
		psi:   0.4,
		omega: 0.2,
	}
}

// Calculate computes the final anomaly risk score Arisk.
// The output is bounded between [0.0, 1.0].
func (c *Calculator) Calculate(vGeo, dVal, iBeh float64) float64 {
	// Clip inputs to range [0.0, 1.0] to prevent mathematical overflows or invalid scores
	vGeo = clamp(vGeo)
	dVal = clamp(dVal)
	iBeh = clamp(iBeh)

	return c.phi*vGeo + c.psi*dVal + c.omega*iBeh
}

// clamp ensures a value resides in the range [0.0, 1.0].
func clamp(val float64) float64 {
	if val < 0.0 {
		return 0.0
	}
	if val > 1.0 {
		return 1.0
	}
	return val
}
