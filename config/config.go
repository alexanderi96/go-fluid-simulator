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
	GameX                   int32
	GameY                   int32
	GameZ                   int32
	TargetFPS               int32
	IsResizable             bool
	ParticleNumber          int32
	ParticleRadius          float32
	ParticleMass            float32
	ParticleInitialSpacing  float32
	ShowVectors             bool
	ScaleFactor             float32
	ParticleElasticity      float32
	WallElasticity          float32
	ApplyGravity            bool
	Gravity                 float32
	ShowQuadtree            bool
	ShowTrail               bool
	ShouldBeProfiled        bool
	UseExperimentalQuadtree bool
	SetRandomRadius         bool
	RadiusMin               float32
	RadiusMax               float32
	SetRandomMass           bool
	MassMin                 float32
	MassMax                 float32
	SetRandomElasticity     bool
	ElasticityMin           float32
	ElasticityMax           float32
	ShowOverlay             bool
	SetRandomColor          bool
	ShowSpeedColor          bool
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
		GameX:                   viper.GetInt32("game_x"),
		GameY:                   viper.GetInt32("game_y"),
		GameZ:                   viper.GetInt32("game_z"),
		TargetFPS:               viper.GetInt32("target_fps"),
		IsResizable:             viper.GetBool("is_resizable"),
		ParticleNumber:          viper.GetInt32("particle_number"),
		ParticleRadius:          float32(viper.GetFloat64("particle_radius")),
		ParticleMass:            float32(viper.GetFloat64("particle_mass")),
		ParticleInitialSpacing:  float32(viper.GetFloat64("particle_initial_spacing")),
		ShowVectors:             viper.GetBool("show_vectors"),
		ScaleFactor:             float32(viper.GetFloat64("scale_factor")),
		ParticleElasticity:      float32(viper.GetFloat64("particle_elasticity")),
		WallElasticity:          float32(viper.GetFloat64("wall_elasticity")),
		ApplyGravity:            viper.GetBool("apply_gravity"),
		Gravity:                 float32(viper.GetFloat64("gravity")),
		ShowQuadtree:            viper.GetBool("show_quadtree"),
		ShowTrail:               viper.GetBool("show_trail"),
		ShouldBeProfiled:        viper.GetBool("should_be_profiled"),
		UseExperimentalQuadtree: viper.GetBool("use_experimental_quadtree"),
		SetRandomRadius:         viper.GetBool("set_random_radius"),
		RadiusMin:               float32(viper.GetFloat64("radius_min")),
		RadiusMax:               float32(viper.GetFloat64("radius_max")),
		SetRandomMass:           viper.GetBool("set_random_mass"),
		MassMin:                 float32(viper.GetFloat64("mass_min")),
		MassMax:                 float32(viper.GetFloat64("mass_max")),
		SetRandomElasticity:     viper.GetBool("set_random_elasticity"),
		ElasticityMin:           float32(viper.GetFloat64("elasticity_min")),
		ElasticityMax:           float32(viper.GetFloat64("elasticity_max")),
		ShowOverlay:             viper.GetBool("show_overlay"),
		SetRandomColor:          viper.GetBool("set_random_color"),
		ShowSpeedColor:          viper.GetBool("show_speed_color"),
	}

	return config, nil
}

func (c *Config) UpdateWindowSettings() {

	currentWidth := int32(rl.GetScreenWidth())
	currentHeight := int32(rl.GetScreenHeight())

	c.WindowWidth = currentWidth
	c.WindowHeight = currentHeight
	c.SidebarWidth = c.WindowWidth / 5

	c.ViewportX = currentWidth - c.SidebarWidth
	c.ViewportY = currentHeight

	c.GameX = c.ViewportX
	c.GameY = c.ViewportY

	if c.FullScreen {
		rl.ToggleFullscreen()
	}
}
