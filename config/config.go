package config

import (
	"github.com/spf13/viper"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Config struct {
	FullScreen              bool
	WindowWidth             int32
	WindowHeight            int32
	SidebarWidth            int32
	ViewportX               int32
	ViewportY               int32
	GameX                   float64
	GameY                   float64
	GameZ                   float64
	TargetFPS               int32
	IsResizable             bool
	UnitNumber              int32
	UnitRadius              float64
	UnitMassMultiplier      float64
	UnitInitialSpacing      float64
	ShowVectors             bool
	ScaleFactor             float64
	UnitElasticity          float64
	WallElasticity          float64
	ApplyGravity            bool
	Gravity                 float64
	ShowOctree              bool
	ShowTrail               bool
	ShouldBeProfiled        bool
	SetRandomRadius         bool
	RadiusMin               float64
	RadiusMax               float64
	SetRandomMassMultiplier bool
	MassMultiplierMin       float64
	MassMultiplierMax       float64
	SetRandomElasticity     bool
	ElasticityMin           float64
	ElasticityMax           float64
	ShowOverlay             bool
	SetRandomColor          bool
	ShowSpeedColor          bool
	UnitsEmitGravity        bool
	UnitRadiusMultiplier    float64
	ShowClusterColor        bool
	OctreeMaxLevel          int8
	MaxUnitNumberPerLevel   int8
	ResolutionSteps         int8
	Frametime               float64
	// TestDuration          float64
	// TestIterations        int32
}

func ReadConfig(filepath string) (*Config, error) {
	viper.SetConfigFile(filepath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		FullScreen:              viper.GetBool("full_screen"),
		WindowWidth:             viper.GetInt32("window_width"),
		WindowHeight:            viper.GetInt32("window_height"),
		GameX:                   float64(viper.GetInt32("game_x")),
		GameY:                   float64(viper.GetInt32("game_y")),
		GameZ:                   float64(viper.GetInt32("game_z")),
		TargetFPS:               viper.GetInt32("target_fps"),
		IsResizable:             viper.GetBool("is_resizable"),
		UnitNumber:              viper.GetInt32("unit_number"),
		UnitRadius:              viper.GetFloat64("unit_radius"),
		UnitMassMultiplier:      viper.GetFloat64("unit_mass_multiplier"),
		UnitInitialSpacing:      viper.GetFloat64("unit_initial_spacing"),
		ShowVectors:             viper.GetBool("show_vectors"),
		ScaleFactor:             viper.GetFloat64("scale_factor"),
		UnitElasticity:          viper.GetFloat64("unit_elasticity"),
		WallElasticity:          viper.GetFloat64("wall_elasticity"),
		ApplyGravity:            viper.GetBool("apply_gravity"),
		Gravity:                 viper.GetFloat64("gravity"),
		ShowOctree:              viper.GetBool("show_octree"),
		ShowTrail:               viper.GetBool("show_trail"),
		ShouldBeProfiled:        viper.GetBool("should_be_profiled"),
		SetRandomRadius:         viper.GetBool("set_random_radius"),
		RadiusMin:               viper.GetFloat64("radius_min"),
		RadiusMax:               viper.GetFloat64("radius_max"),
		SetRandomMassMultiplier: viper.GetBool("set_random_mass_multiplier"),
		MassMultiplierMin:       viper.GetFloat64("mass_multiplier_min"),
		MassMultiplierMax:       viper.GetFloat64("mass_multiplier_max"),
		SetRandomElasticity:     viper.GetBool("set_random_elasticity"),
		ElasticityMin:           viper.GetFloat64("elasticity_min"),
		ElasticityMax:           viper.GetFloat64("elasticity_max"),
		ShowOverlay:             viper.GetBool("show_overlay"),
		SetRandomColor:          viper.GetBool("set_random_color"),
		ShowSpeedColor:          viper.GetBool("show_speed_color"),
		UnitsEmitGravity:        viper.GetBool("units_emit_gravity"),
		UnitRadiusMultiplier:    viper.GetFloat64("unit_radius_multiplier"),
		OctreeMaxLevel:          int8(viper.GetInt("octree_max_level")),
		MaxUnitNumberPerLevel:   int8(viper.GetInt("max_unit_number_per_level")),
		ResolutionSteps:         int8(viper.GetInt("resolution_steps")),
		Frametime:               viper.GetFloat64("frametime"),
		// TestDuration:          viper.GetFloat64("test_duration"),
		// TestIterations:        viper.GetInt32("test_iterations"),
	}

	return config, nil
}

func (c *Config) UpdateWindowSettings() {

	currentWidth := int32(rl.GetScreenWidth())
	currentHeight := int32(rl.GetScreenHeight())

	c.WindowWidth = currentWidth
	c.WindowHeight = currentHeight

	if c.FullScreen {
		rl.ToggleFullscreen()
	}
}

func (c *Config) ResizeViewport(X, Y int32) {
	c.ViewportX = c.WindowWidth + X
	c.ViewportY = c.WindowHeight + Y

	// c.GameX = c.ViewportX
	// c.GameY = c.ViewportY
}
