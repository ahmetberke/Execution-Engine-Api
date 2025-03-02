package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func InitLogger() {
	// JSON formatında loglama
	Log.SetFormatter(&logrus.JSONFormatter{})

	// Log seviyesini INFO olarak ayarla
	Log.SetLevel(logrus.InfoLevel)

	// Logları bir dosyaya yaz
	file, err := os.OpenFile("execution.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Warn("Failed to log to file, using default stderr")
		return
	}

	// Logları hem ekrana (stdout) hem de dosyaya yazdır
	multiWriter := io.MultiWriter(os.Stdout, file)
	Log.SetOutput(multiWriter)

	Log.Info("Logger initialized successfully")
}
