package config

type Config struct {
	Port         int `env:"PORT" envDefault:"8000"`
	GameTickRate int `env:"GAME_TICK_RATE" envDefault:"30"`
}
