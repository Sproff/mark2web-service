package web

import (
	"net"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/thealamu/mark2web-service/internal/pkg/db"
	"github.com/thealamu/mark2web-service/internal/pkg/mark2web"
)

func Start() int {
	srv, err := newServer(getRunAddr(), logger, service)
	if err != nil {
		log.Error(err)
		return 1
	}
	log.Infof("starting server on %s\n", srv.Addr)
	// TODO(thealamu): Implement graceful shutdown on interrupt
	if err := srv.ListenAndServe(); err != nil {
		log.Error(err)
		return 1
	}
	log.Traceln("server stopped")
	return 0
}

func service(s *server) error {
	logger := func(srvc *mark2web.Service) error {
		// use the server's logger in service
		srvc.Logger = s.logger
		return nil
	}

	var database func(srvc *mark2web.Service) error
	if hasEnv("GOOGLE_APPLICATION_CREDENTIALS") {
		// Use firebase
		database = func(srvc *mark2web.Service) error {
			firebaseDB, err := db.NewFirebaseDB(srvc.Logger)
			if err != nil {
				return err
			}
			srvc.DB = firebaseDB
			return nil
		}
	} else {
		database = func(srvc *mark2web.Service) error {
			srvc.DB = &db.FSDatabase{
				BaseDir: ".",
			}
			return nil
		}
	}

	srvc, err := mark2web.NewService(logger, database)
	if err != nil {
		return err
	}

	s.service = srvc
	return nil
}

// logger offers a suitable logger for use in handlers
func logger(s *server) error {
	logLevel, err := log.ParseLevel(getLogLevelFromEnv())
	if err != nil {
		logLevel = log.InfoLevel
	}
	s.logger = &log.Logger{
		Out: os.Stderr,
		Formatter: &log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		},
		Level: logLevel,
	}
	return nil
}

// func service(l *log.Logger) *mark2web.Service {
// 	var dbImpl db.DB
// 	dbImpl, err := db.NewFirebaseDB(l)
// 	if err != nil {
// 		l.Error(err)
// 		dbImpl = &db.FSDatabase{
// 			// TODO(thealamu): Use suitable system directory to store local files
// 			BaseDir: ".",
// 		}
// 	}
// 	return &mark2web.Service{DB: dbImpl}
// }

// // httpServer returns a simple, configured http server
// func httpServer() *http.Server {
// 	return &http.Server{
// 		Addr:         getRunAddr(),
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 5 * time.Second,
// 	}
// }

// getRunAddr returns the address to start the server on.
// If no port in environment, it defaults to 8080.
func getRunAddr() string {
	port := getPortFromEnv()
	if port == "" {
		port = "8080"
	}
	return net.JoinHostPort("", port)
}

// getLastPath returns the last path item in a URL.
// For example, for the URL https://example.com/12345, it returns 12345.
// For a URL with no path, it returns an empty string.
func getLastPath(URL string) string {
	urlObj, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	path := urlObj.Path
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return ""
	}
	return path[i+1:]
}
