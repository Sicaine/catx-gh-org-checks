package main

import (
	"encoding/json"
	"github.com/catena-x/gh-org-checks/pkg/data"
	"github.com/catena-x/gh-org-checks/pkg/testers"
	"github.com/catena-x/gh-org-checks/pkg/testrunner"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	orgReport  = data.OrgReports{}
	testRunner *testrunner.TestRunner
)

// ref: https://github.com/gorilla/mux#serving-single-page-applications
// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {

	log.Printf("Starting service ...")
	setLogLevel()

	initTestSuiteAndSchedule(*testrunner.NewTestRunner())

	router := mux.NewRouter()
	router.HandleFunc("/report", returnOrgReport).Methods(http.MethodGet)

	spa := spaHandler{staticPath: "./dashboard/dist/dashboard", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	log.Fatal(http.ListenAndServe(":8000", router))

}

func initTestSuiteAndSchedule(testRunner testrunner.TestRunner) {
	testRunner.AddToTestSuites(testers.NewReadMeTester)
	testRunner.AddToTestSuites(testers.NewHelmChartTester)
	testRunner.AddToTestSuites(testers.NewReleaseTester)
	testRunner.AddToTestSuites(testers.NewOSSTester)
	testRunner.AddToTestSuites(testers.NewSecurityActionTester)

	scheduleCronJobs(testRunner)
}

func setLogLevel() {
	log.SetLevel(log.DebugLevel)
}

func scheduleCronJobs(testRunner testrunner.TestRunner) {
	log.Println("scheduled test cronjob")
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().Do(func() {
		go updateTestReport(testRunner)
	},
	)

	s.StartAsync()
}

func updateTestReport(testRunner testrunner.TestRunner) {
	log.Println("update test report")
	orgReport = testRunner.PerformRepoChecks()
}

func returnOrgReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("returning test report")
	json.NewEncoder(w).Encode(orgReport)
}
