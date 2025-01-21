package constants

// Costanti fisiche universali
const (
	// Costante gravitazionale universale (m³/kg⋅s²)
	G = 6.67430e-11

	// Massa dei corpi celesti (kg)
	EarthMass = 5.972e24
	MoonMass  = 7.34767309e22
	SunMass   = 1.989e30

	// Raggi dei corpi celesti (m)
	EarthRadius = 6.371e6
	MoonRadius  = 1.737e6
	SunRadius   = 6.957e8

	// Densità medie (kg/m³)
	EarthDensity = 5510
	MoonDensity  = 3340
	SunDensity   = 1410

	// Costanti termiche
	StefanBoltzmannConstant = 5.67e-8 // Costante di Stefan-Boltzmann (W/m²⋅K⁴)
	AmbientTemperature      = 20.0    // Temperatura ambiente (°C)
	MaxHeatTransferDistance = 10.0    // Distanza massima per il trasferimento di calore (m)
	CoolingRate             = 0.01    // Tasso di raffreddamento

	// Costanti di rendering
	Segments = 10 // Numero di segmenti per la mesh sferica
)
