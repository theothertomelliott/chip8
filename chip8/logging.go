package chip8

import "log"

type logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type loggerWithToggle struct {
	Enabled bool
}

func (l *loggerWithToggle) Printf(format string, v ...interface{}) {
	if l.Enabled {
		log.Printf(format, v...)
	}
}

func (l *loggerWithToggle) Println(v ...interface{}) {
	if l.Enabled {
		log.Println(v...)
	}
}

type logging struct {
	Opcodes *loggerWithToggle
}
