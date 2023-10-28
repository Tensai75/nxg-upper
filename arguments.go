package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	parser "github.com/alexflint/go-arg"
)

// arguments structure
type Args struct {
	Path               string `arg:"positional,required" help:"Path to be uploaded" placeholder:"PATH"`
	Host               string `arg:"--host" help:"Usenet server host name or IP address" placeholder:"HOST"`
	Port               int    `arg:"--port" help:"Usenet server port number" placeholder:"INT"`
	SSL                bool   `arg:"-"`
	SSL_arg            string `arg:"--ssl" help:"Use SSL" placeholder:"true|false"`
	NntpUser           string `arg:"--user" help:"Username to connect to the usenet server" placeholder:"STRING"`
	NntpPass           string `arg:"--pass" help:"Password to connect to the usenet server" placeholder:"STRING"`
	Connections        int    `arg:"--connections" help:"Number of connections to use to connect to the usenet server" placeholder:"INT"`
	ConnRetries        int    `arg:"--connretries" help:"Number of retries upon connection error" placeholder:"INT"`
	ConnWaitTime       int    `arg:"--connwaittime" help:"Time to wait in seconds before trying to re-connect" placeholder:"INT"`
	Groups             string `arg:"--groups" help:"List of groups (separated by commas) to post to" placeholder:"GROUPS"`
	LineLength         int    `arg:"--linelength" help:"Line length of the yEnc encoded article body" placeholder:"INT"`
	ArticleSize        int64  `arg:"--articlesize" help:"Size of the article body in bytes" placeholder:"BYTES"`
	Retries            int    `arg:"--retries" help:"Number of retries before article posting fails" placeholder:"INT"`
	Poster             string `arg:"--poster" help:"Poster (From address) for the articles (leave empty for random poster)" placeholder:"EMAIL"`
	MakeRar            bool   `arg:"-"`
	MakeRar_arg        string `arg:"--rar" help:"Make rar archive" placeholder:"true|false"`
	MakeVolumes        bool   `arg:"-"`
	MakeVolumes_arg    string `arg:"--makevolumes" help:"Create rar volumes" placeholder:"true|false"`
	MaxVolumes         int64  `arg:"--maxvolumes" help:"Maximum amount of volumes" placeholder:"INT"`
	VolumeSize         int64  `arg:"--volumesize" help:"Minimum volume size in bytes" placeholder:"BYTES"`
	Encrypt            bool   `arg:"-"`
	Encrypt_arg        string `arg:"--encrypt" help:"Encrypt the rar file with a password" placeholder:"true|false"`
	Password           string `arg:"--password" help:"Password for the rar file" placeholder:"STRING"`
	Compression        int    `arg:"--compression" help:"Compression level for rar file" placeholder:"0-9"`
	RarExe             string `arg:"--rarexe" help:"Path to the rar executable" placeholder:"PATH"`
	MakePar2           bool   `arg:"-"`
	MakePar2_arg       string `arg:"--par2" help:"Make par2 files" placeholder:"true|false"`
	Redundancy         int    `arg:"--redundancy" help:"Redundancy in %" placeholder:"0-100"`
	Par2Exe            string `arg:"--parexe" help:"Path to the par executable (either par2 [par2cmdline] or parpar)" placeholder:"PATH"`
	CsvPath            string `arg:"--csvpath" help:"CSV file path to log title, header, password, groups and date (leave empty too disable)" placeholder:"PATH"`
	CsvDelimiter       string `arg:"--csvdelimiter" help:"CSV file delimiter" placeholder:"STRING"`
	DelTempFolder      bool   `arg:"-"`
	DelTempFolder_arg  string `arg:"--deltemp" help:"Delete temp folder at program end" placeholder:"true|false"`
	DelInputFolder     bool   `arg:"-"`
	DelInputFolder_arg string `arg:"--delinput" help:"Delete input folder after successful upload" placeholder:"true|false"`
	TempPath           string `arg:"--tmp" help:"Temporary path for rar/par2" placeholder:"PATH"`
	LogFilePath        string `arg:"--log" help:"Path where to save the log file" placeholder:"PATH"`
	NzbPath            string `arg:"--nzb" help:"Path where to save the NZB file" placeholder:"PATH"`
	//	FaildArticles      string   `arg:"--faildarticles" help:"Path where to save failed articles"`
	Verbose     int      `arg:"--verbose" help:"Verbosity level of cmd line output"  placeholder:"0-3"`
	Debug       bool     `arg:"-"`
	Debug_arg   string   `arg:"--debug" help:"Activate debug mode" placeholder:"true|false"`
	Test        string   `arg:"--test" help:"Activate test mode and save messages to PATH" placeholder:"PATH"`
	GroupsArray []string `arg:"-"`
}

// version information
func (Args) Version() string {
	return "\n" + appName + " " + appVersion + "\n"
}

// additional description
func (Args) Epilogue() string {
	return "\nParameters that are passed as arguments have precedence over the settings in the configuration file.\n"
}

// parser variable
var argParser *parser.Parser

func parseArguments() {

	parserConfig := parser.Config{
		IgnoreEnv: true,
	}

	// parse flags
	argParser, _ = parser.NewParser(parserConfig, &conf)
	if err := parser.Parse(&conf); err != nil {
		if err.Error() == "help requested by user" {
			writeHelp(argParser)
			fmt.Println(conf.Epilogue())
			exit(0, nil)
		} else if err.Error() == "version requested by user" {
			fmt.Println(conf.Version())
			exit(0, nil)
		}
		writeUsage(argParser)
		checkForFatalErr(err)
	}

}

func checkArguments() {
	if conf.Groups != "" {
		groupString := strings.Split(conf.Groups, ",")
		for _, group := range groupString {
			if group != "" {
				conf.GroupsArray = append(conf.GroupsArray, strings.Replace(strings.TrimSpace(group), "a.b.", "alt.binaries.", 1))
			}
		}
	} else {
		writeUsage(argParser)
		checkForFatalErr(fmt.Errorf("No groups provided"))
	}

	if conf.Poster == "" {
		conf.Poster = randomString(12, 0, true) + "@" + randomString(8, 1, false) + "." + randomString(3, 2, false)
	}

	if conf.TempPath == "" {
		conf.TempPath = os.TempDir()
	}

	if !filepath.IsAbs(conf.TempPath) {
		if conf.TempPath, err = filepath.Abs(filepath.Join(homePath, conf.TempPath)); err != nil {
			checkForFatalErr(fmt.Errorf("Unable to determine temporary path: %v", err))
		}
	}
	if !filepath.IsAbs(conf.NzbPath) {
		if conf.NzbPath, err = filepath.Abs(filepath.Join(homePath, conf.NzbPath)); err != nil {
			checkForFatalErr(fmt.Errorf("Unable to determine NZB file path: %v", err))
		}
	}
	/*
		if !filepath.IsAbs(conf.FaildArticles) {
			if conf.FaildArticles, err = filepath.Abs(filepath.Join(homePath, conf.FaildArticles)); err != nil {
				checkForFatalErr(fmt.Errorf("Unable to determine failed articles path: %v", err))
			}
		}
	*/
	// check bools
	if conf.SSL_arg != "" {
		if conf.SSL_arg == "true" {
			conf.SSL = true
		} else if conf.SSL_arg == "false" {
			conf.SSL = false
		}
	}
	if conf.MakeRar_arg != "" {
		if conf.MakeRar_arg == "true" {
			conf.MakeRar = true
		} else if conf.MakeRar_arg == "false" {
			conf.MakeRar = false
		}
	}
	if conf.MakeVolumes_arg != "" {
		if conf.MakeVolumes_arg == "true" {
			conf.MakeVolumes = true
		} else if conf.MakeVolumes_arg == "false" {
			conf.MakeVolumes = false
		}
	}
	if conf.Encrypt_arg != "" {
		if conf.Encrypt_arg == "true" {
			conf.Encrypt = true
		} else if conf.Encrypt_arg == "false" {
			conf.Encrypt = false
		}
	}
	if conf.MakePar2_arg != "" {
		if conf.MakePar2_arg == "true" {
			conf.MakePar2 = true
		} else if conf.MakePar2_arg == "false" {
			conf.MakePar2 = false
		}
	}
	if conf.DelTempFolder_arg != "" {
		if conf.DelTempFolder_arg == "true" {
			conf.DelTempFolder = true
		} else if conf.DelTempFolder_arg == "false" {
			conf.DelTempFolder = false
		}
	}
	if conf.DelInputFolder_arg != "" {
		if conf.DelInputFolder_arg == "true" {
			conf.DelInputFolder = true
		} else if conf.DelInputFolder_arg == "false" {
			conf.DelInputFolder = false
		}
	}
	if conf.Debug_arg != "" {
		if conf.Debug_arg == "true" {
			conf.Debug = true
		} else if conf.Debug_arg == "false" {
			conf.Debug = false
		}
	}
}

func writeUsage(parser *parser.Parser) {
	var buf bytes.Buffer
	parser.WriteUsage(&buf)
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

}

func writeHelp(parser *parser.Parser) {
	var buf bytes.Buffer
	parser.WriteHelp(&buf)
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

}
