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
	Path                string `arg:"positional,required" help:"Path to be uploaded" placeholder:"PATH"`
	Host                string `arg:"--host" help:"Usenet server host name or IP address" placeholder:"HOST"`
	Port                uint32 `arg:"--port" help:"Usenet server port number" placeholder:"INT"`
	SSL                 bool   `arg:"-"`
	SSL_arg             string `arg:"--ssl" help:"Use SSL" placeholder:"true|false"`
	NntpUser            string `arg:"--user" help:"Username to connect to the usenet server" placeholder:"STRING"`
	NntpPass            string `arg:"--pass" help:"Password to connect to the usenet server" placeholder:"STRING"`
	Connections         uint32 `arg:"--connections" help:"Number of connections to use to connect to the usenet server" placeholder:"INT"`
	ConnRetries         uint32 `arg:"--connretries" help:"Number of retries upon connection error" placeholder:"INT"`
	ConnWaitTime        uint32 `arg:"--connwaittime" help:"Time to wait in seconds before trying to re-connect" placeholder:"INT"`
	Groups              string `arg:"--groups" help:"List of groups (separated by commas) to post to" placeholder:"GROUPS"`
	LineLength          uint32 `arg:"--linelength" help:"Line length of the yEnc encoded article body" placeholder:"INT"`
	ArticleSize         uint64 `arg:"--articlesize" help:"Size of the article body in bytes" placeholder:"BYTES"`
	Retries             uint32 `arg:"--retries" help:"Number of retries before article posting fails" placeholder:"INT"`
	Poster              string `arg:"--poster" help:"Poster (From address) for the articles (leave empty for random poster for each upload)" placeholder:"EMAIL"`
	Obfuscate           bool   `arg:"-"`
	Obfuscate_arg       string `arg:"--obfuscate" help:"Obfuscate the subject of upload" placeholder:"true|false"`
	ObfuscatePoster     bool   `arg:"-"`
	ObfuscatePoster_arg string `arg:"--obfuscateposter" help:"Obfuscate also the poster (From address)" placeholder:"true|false"`
	ObfuscateYenc       bool   `arg:"-"`
	ObfuscateYenc_arg   string `arg:"--obfuscateyenc" help:"Obfuscate also the yenc header" placeholder:"true|false"`
	HeaderCheck         bool   `arg:"-"`
	HeaderCheck_arg     string `arg:"--headercheck" help:"Activate header check" placeholder:"true|false"`
	HeaderCheckDelay    uint32 `arg:"--headercheckdelay" help:"Header check delay in seconds" placeholder:"INT"`
	HeaderCheckConns    uint32 `arg:"--headercheckconns" help:"Header check connections" placeholder:"INT"`
	MakeRar             bool   `arg:"-"`
	MakeRar_arg         string `arg:"--rar" help:"Make rar archive" placeholder:"true|false"`
	ObfuscateRar        bool   `arg:"-"`
	ObfuscateRar_arg    string `arg:"--obfuscaterar" help:"Obfuscate rar archive name" placeholder:"true|false"`
	MakeVolumes         bool   `arg:"-"`
	MakeVolumes_arg     string `arg:"--makevolumes" help:"Create rar volumes" placeholder:"true|false"`
	MaxVolumes          uint64 `arg:"--maxvolumes" help:"Maximum amount of volumes" placeholder:"INT"`
	VolumeSize          uint64 `arg:"--volumesize" help:"Minimum volume size in bytes" placeholder:"BYTES"`
	Encrypt             bool   `arg:"-"`
	Encrypt_arg         string `arg:"--encrypt" help:"Encrypt the rar file with a password" placeholder:"true|false"`
	Password            string `arg:"--password" help:"Password for the rar file" placeholder:"STRING"`
	PasswordLength      uint32 `arg:"--passwordlength" help:"Length of the random password for the rar file" placeholder:"NUMBER"`
	Compression         uint32 `arg:"--compression" help:"Compression level for rar file" placeholder:"0-9"`
	RarExe              string `arg:"--rarexe" help:"Path to the rar executable" placeholder:"PATH"`
	MakePar2            bool   `arg:"-"`
	MakePar2_arg        string `arg:"--par2" help:"Make par2 files" placeholder:"true|false"`
	Redundancy          uint32 `arg:"--redundancy" help:"Redundancy in %" placeholder:"0-100"`
	Par2Exe             string `arg:"--parexe" help:"Path to the par executable (either par2 [par2cmdline] or parpar)" placeholder:"PATH"`
	CsvPath             string `arg:"--csvpath" help:"CSV file path to log title, header, password, groups and date (leave empty too disable)" placeholder:"PATH"`
	CsvDelimiter        string `arg:"--csvdelimiter" help:"CSV file delimiter" placeholder:"STRING"`
	DelTempFolder       bool   `arg:"-"`
	DelTempFolder_arg   string `arg:"--deltemp" help:"Delete temp folder at program end" placeholder:"true|false"`
	DelInputFolder      bool   `arg:"-"`
	DelInputFolder_arg  string `arg:"--delinput" help:"Delete input folder after successful upload" placeholder:"true|false"`
	TempPath            string `arg:"--tmp" help:"Temporary path for rar/par2" placeholder:"PATH"`
	LogFilePath         string `arg:"--log" help:"Path where to save the log file" placeholder:"PATH"`
	NzbPath             string `arg:"--nzb" help:"Path where to save the NZB file" placeholder:"PATH"`
	//	FaildArticles      string   `arg:"--faildarticles" help:"Path where to save failed articles"`
	Verbose     uint32   `arg:"--verbose" help:"Verbosity level of cmd line output"  placeholder:"0-3"`
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
	var err error

	if conf.Groups != "" {
		groupString := strings.Split(conf.Groups, ",")
		for _, group := range groupString {
			if group != "" {
				conf.GroupsArray = append(conf.GroupsArray, strings.Replace(strings.TrimSpace(group), "a.b.", "alt.binaries.", 1))
			}
		}
	} else {
		writeUsage(argParser)
		checkForFatalErr(fmt.Errorf("no groups provided"))
	}

	if conf.Poster == "" {
		conf.Poster = randomString(12, 0, true) + "@" + randomString(8, 1, false) + "." + randomString(3, 2, false)
	}

	if conf.TempPath == "" {
		conf.TempPath = os.TempDir()
	}

	if !filepath.IsAbs(conf.TempPath) {
		if conf.TempPath, err = filepath.Abs(filepath.Join(homePath, conf.TempPath)); err != nil {
			checkForFatalErr(fmt.Errorf("unable to determine temporary path: %v", err))
		}
	}
	if !filepath.IsAbs(conf.NzbPath) {
		if conf.NzbPath, err = filepath.Abs(filepath.Join(homePath, conf.NzbPath)); err != nil {
			checkForFatalErr(fmt.Errorf("unable to determine NZB file path: %v", err))
		}
	}
	if !filepath.IsAbs(conf.LogFilePath) {
		if conf.LogFilePath, err = filepath.Abs(filepath.Join(homePath, conf.LogFilePath)); err != nil {
			checkForFatalErr(fmt.Errorf("unable to determine log file path: %v", err))
		}
	}
	if !filepath.IsAbs(conf.CsvPath) {
		if conf.CsvPath, err = filepath.Abs(filepath.Join(homePath, conf.CsvPath)); err != nil {
			checkForFatalErr(fmt.Errorf("unable to determine csv file path: %v", err))
		}
	}
	if conf.PasswordLength == 0 {
		conf.PasswordLength = 25
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
	if conf.ObfuscateRar_arg != "" {
		if conf.ObfuscateRar_arg == "true" {
			conf.ObfuscateRar = true
		} else if conf.ObfuscateRar_arg == "false" {
			conf.ObfuscateRar = false
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
	if conf.Obfuscate_arg != "" {
		if conf.Obfuscate_arg == "true" {
			conf.Obfuscate = true
		} else if conf.Obfuscate_arg == "false" {
			conf.Obfuscate = false
		}
	}
	if conf.ObfuscatePoster_arg != "" {
		if conf.ObfuscatePoster_arg == "true" {
			conf.ObfuscatePoster = true
		} else if conf.ObfuscatePoster_arg == "false" {
			conf.ObfuscatePoster = false
		}
	}
	if conf.ObfuscateYenc_arg != "" {
		if conf.ObfuscateYenc_arg == "true" {
			conf.ObfuscateYenc = true
		} else if conf.ObfuscateYenc_arg == "false" {
			conf.ObfuscateYenc = false
		}
	}
	if conf.HeaderCheck_arg != "" {
		if conf.HeaderCheck_arg == "true" {
			conf.HeaderCheck = true
		} else if conf.HeaderCheck_arg == "false" {
			conf.HeaderCheck = false
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
