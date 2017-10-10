package operator

// Config is the job resource config.
type Config struct {
	MetaData struct {
		Name string
	}

	Spec struct {
		Trainer struct {
			Entrypoint  string
			Workspace   string
			MinInstance int
			MaxInstance int
			Resources   struct {
				Limits struct {
					GPU int
					CPU float64
					Mem float64
				}
				Requests struct {
					GPU int
					CPU float64
					Mem float64
				}
			}
		}

		Pserver struct {
			MinInstance int
			Resources   struct {
				Limits struct {
					CPU float64
					Mem float64
				}
				Requests struct {
					GPU int
					CPU float64
					Mem float64
				}
			}
		}

		Master struct {
			Resources struct {
				Limits struct {
					CPU float64
					Mem float64
				}
				Requests struct {
					GPU int
					CPU float64
					Mem float64
				}
			}
		}
	}
}

func (s *Config) NeedGPU() bool {
	return s.Spec.Trainer.Resources.Limits.GPU > 0
}

func (s *Config) Elastic() bool {
	return s.Spec.Trainer.MinInstance != s.Spec.Trainer.MaxInstance
}
