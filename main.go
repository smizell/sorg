package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brandur/modulir"
	"github.com/joeshaw/envdecode"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Main
//
//
//
//////////////////////////////////////////////////////////////////////////////

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sorg",
		Short: "Sorg is a static site generator",
		Long: strings.TrimSpace(`
Sorg is a static site generator for Brandur's personal
homepage and some of its adjacent functions. See the product
in action at https://brandur.org.`),
	}

	buildCommand := &cobra.Command{
		Use:   "build",
		Short: "Run a single build loop",
		Long: strings.TrimSpace(`
Starts the build loop that watches for local changes and runs
when they're detected. A webserver is started on PORT (default
5002).`),
		Run: func(cmd *cobra.Command, args []string) {
			modulir.Build(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(buildCommand)

	loopCommand := &cobra.Command{
		Use:   "loop",
		Short: "Start build and serve loop",
		Long: strings.TrimSpace(`
Runs the build loop one time and places the result in TARGET_DIR
(default ./public/).`),
		Run: func(cmd *cobra.Command, args []string) {
			modulir.BuildLoop(getModulirConfig(), build)
		},
	}
	rootCmd.AddCommand(loopCommand)

	var live bool
	var staging bool
	passagesCommand := &cobra.Command{
		Use:   "passages [source .md file]",
		Short: "Email a Passages newsletter",
		Long: strings.TrimSpace(`
Emails the Passages newsletter at the location given as argument.
Note that MAILGUN_API_KEY must be set in the environment for this
to work as it executes against the Mailgun API.`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := &modulir.Context{Log: getLog()}
			sendPassages(c, args[0], live, staging)
		},
	}
	passagesCommand.Flags().BoolVar(&live, "live", false,
		"Send to list (as opposed to dry run)")
	passagesCommand.Flags().BoolVar(&staging, "staging", false,
		"Send to staging list (as opposed to dry run)")
	rootCmd.AddCommand(passagesCommand)

	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v", err)
		os.Exit(1)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Conf contains configuration information for the command. It's extracted from
// environment variables.
type Conf struct {
	// AbsoluteURL is the absolute URL where the compiled site will be hosted.
	// It's used for things like Atom feeds and sending email.
	AbsoluteURL string `env:"ABSOLUTE_URL,default=https://brandur.org"`

	// BlackSwanDatabaseURL is a connection string for a database to connect to
	// in order to extract books, tweets, runs, etc.
	BlackSwanDatabaseURL string `env:"BLACK_SWAN_DATABASE_URL"`

	// Concurrency is the number of build Goroutines that will be used to
	// perform build work items.
	Concurrency int `env:"CONCURRENCY,default=30"`

	// Drafts is whether drafts of articles and fragments should be compiled
	// along with their published versions.
	//
	// Activating drafts also prompts the creation of a robots.txt to make sure
	// that drafts aren't inadvertently accessed by web crawlers.
	Drafts bool `env:"DRAFTS,default=false"`

	// GoogleAnalyticsID is the account identifier for Google Analytics to use.
	GoogleAnalyticsID string `env:"GOOGLE_ANALYTICS_ID"`

	// LocalFonts starts using locally downloaded versions of Google Fonts.
	// This is not ideal for real deployment because you won't be able to
	// leverage Google's CDN and the caching that goes with it, and may not get
	// the font format for requesting browsers, but good for airplane rides
	// where you otherwise wouldn't have the fonts.
	LocalFonts bool `env:"LOCAL_FONTS,default=false"`

	// MailgunAPIKey is a key for Mailgun used to send email. It's required
	// when using the `passages` command.
	MailgunAPIKey string `env:"MAILGUN_API_KEY"`

	// NumAtomEntries is the number of entries to put in Atom feeds.
	NumAtomEntries int `env:"NUM_ATOM_ENTRIES,default=20"`

	// Port is the port on which to serve HTTP when looping in development.
	Port int `env:"PORT,default=5002"`

	// TargetDir is the target location where the site will be built to.
	TargetDir string `env:"TARGET_DIR,default=./public"`

	// Verbose is whether the program will print debug output as it's running.
	Verbose bool `env:"VERBOSE,default=false"`
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

func getLog() modulir.LoggerInterface {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	return log
}

// getModulirConfig interprets Conf to produce a configuration suitable to pass
// to a Modulir build loop.
func getModulirConfig() *modulir.Config {

	return &modulir.Config{
		Concurrency: conf.Concurrency,
		Log:         getLog(),
		Port:        conf.Port,
		SourceDir:   ".",
		TargetDir:   conf.TargetDir,
	}
}
