package config

import (
	"github.com/spf13/viper"
)

// Config rappresenta la configurazione dell'applicazione
type Config struct {
	WindowWidth    int32
	WindowHeight   int32
	SidebarWidth   int32
	GameWidth      int32
	TargetFPS      int32
	IsResizable    bool
	ParticleNumber int32
	ParticleRadius float32
	ParticleMass   float32
}

// ReadConfig legge il file di configurazione e restituisce un'istanza di Config
func ReadConfig(filepath string) (*Config, error) {
	viper.SetConfigFile(filepath)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		WindowWidth:    viper.GetInt32("window_width"),
		WindowHeight:   viper.GetInt32("window_height"),
		TargetFPS:      viper.GetInt32("target_fps"),
		IsResizable:    viper.GetBool("is_resizable"),
		ParticleNumber: viper.GetInt32("particle_number"),
		ParticleRadius: float32(viper.GetFloat64("particle_radius")),
		ParticleMass:   float32(viper.GetFloat64("particle_mass")),
	}

	config.SidebarWidth = config.WindowWidth / 5
	config.GameWidth = config.WindowWidth - config.SidebarWidth

	return config, nil
}
