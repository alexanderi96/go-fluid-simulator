package config

import (
	"github.com/spf13/viper"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Config struct {
	FullScreen                  bool
	WindowWidth                 int32
	WindowHeight                int32
	SidebarWidth                int32
	ViewportX                   int32
	ViewportY                   int32
	GameX                       int32
	GameY                       int32
	GameZ                       int32
	TargetFPS                   int32
	IsResizable                 bool
	UnitNumber                  int32
	UnitRadius                  float32
	UnitMassMultiplier          float32
	UnitInitialSpacing          float32
	ShowVectors                 bool
	ScaleFactor                 float32
	UnitElasticity              float32
	WallElasticity              float32
	ApplyGravity                bool
	Gravity                     float32
	ShowQuadtree                bool
	ShowTrail                   bool
	ShouldBeProfiled            bool
	UseExperimentalQuadtree     bool
	SetRandomRadius             bool
	RadiusMin                   float32
	RadiusMax                   float32
	SetRandomMassMultiplier     bool
	MassMultiplierMin           float32
	MassMultiplierMax           float32
	SetRandomElasticity         bool
	ElasticityMin               float32
	ElasticityMax               float32
	ShowOverlay                 bool
	SetRandomColor              bool
	ShowSpeedColor              bool
	UnitsEmitGravity            bool
	UnitGravitationalMultiplier float32
}

func ReadConfig(filepath string) (*Config, error) {
	viper.SetConfigFile(filepath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		FullScreen:                  viper.GetBool("full_screen"),
		WindowWidth:                 viper.GetInt32("window_width"),
		WindowHeight:                viper.GetInt32("window_height"),
		GameX:                       viper.GetInt32("game_x"),
		GameY:                       viper.GetInt32("game_y"),
		GameZ:                       viper.GetInt32("game_z"),
		TargetFPS:                   viper.GetInt32("target_fps"),
		IsResizable:                 viper.GetBool("is_resizable"),
		UnitNumber:                  viper.GetInt32("unit_number"),
		UnitRadius:                  float32(viper.GetFloat64("unit_radius")),
		UnitMassMultiplier:          float32(viper.GetFloat64("unit_mass_multiplier")),
		UnitInitialSpacing:          float32(viper.GetFloat64("unit_initial_spacing")),
		ShowVectors:                 viper.GetBool("show_vectors"),
		ScaleFactor:                 float32(viper.GetFloat64("scale_factor")),
		UnitElasticity:              float32(viper.GetFloat64("unit_elasticity")),
		WallElasticity:              float32(viper.GetFloat64("wall_elasticity")),
		ApplyGravity:                viper.GetBool("apply_gravity"),
		Gravity:                     float32(viper.GetFloat64("gravity")),
		ShowQuadtree:                viper.GetBool("show_quadtree"),
		ShowTrail:                   viper.GetBool("show_trail"),
		ShouldBeProfiled:            viper.GetBool("should_be_profiled"),
		UseExperimentalQuadtree:     viper.GetBool("use_experimental_quadtree"),
		SetRandomRadius:             viper.GetBool("set_random_radius"),
		RadiusMin:                   float32(viper.GetFloat64("radius_min")),
		RadiusMax:                   float32(viper.GetFloat64("radius_max")),
		SetRandomMassMultiplier:     viper.GetBool("set_random_mass_multiplier"),
		MassMultiplierMin:           float32(viper.GetFloat64("mass_multiplier_min")),
		MassMultiplierMax:           float32(viper.GetFloat64("mass_multiplier_max")),
		SetRandomElasticity:         viper.GetBool("set_random_elasticity"),
		ElasticityMin:               float32(viper.GetFloat64("elasticity_min")),
		ElasticityMax:               float32(viper.GetFloat64("elasticity_max")),
		ShowOverlay:                 viper.GetBool("show_overlay"),
		SetRandomColor:              viper.GetBool("set_random_color"),
		ShowSpeedColor:              viper.GetBool("show_speed_color"),
		UnitsEmitGravity:            viper.GetBool("units_emit_gravity"),
		UnitGravitationalMultiplier: float32(viper.GetFloat64("unit_gravitational_multiplier")),
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

	c.GameX = c.ViewportX
	c.GameY = c.ViewportY
}
