package main

func defaultConfig() string {
	return `# Usenet server settings
# Usenet server host name or IP address
Host: "news.newshosting.com"
# Usenet server port number
Port: 119
# Use SSL if set to true
SSL: false
# Username to connect to the usenet server
NntpUser: ""
# Password to connect to the usenet server
NntpPass: ""
# Number of connections to use to connect to the usenet server
Connections: 50
# Number of retries upon connection error
ConnRetries: 3
# Time to wait in seconds before trying to re-connect
ConnWaitTime: 30

# Posting settings
# List of groups (separated by commas) to post to
Groups: "a.b.test"
# Line length of yEnc encoded body of the article
LineLength: 128
# Size of the article body in bytes (will be adjusted to the nearest multiple of 64)
ArticleSize: 768000
# Poster (From address) for the articles (leave empty for random poster)
Poster: ""
# Obfuscate the upload
Obfuscate: true
# Number of retries before article posting fails
Retries: 3

# RAR settings
# Make rar archive if set to true
MakeRar: true
# Encrypt rar file with a password if set to true. If no password is provided with flag --password a random password is generated
# If set to false, the rar file will nevertheless be encrypted if flag --password is provided
Encrypt: true
# Length of the random password
PasswordLength: 25
# Compression level from 0 to 9 (0 = no compression / 9 = max. compression)
Compression: 0
# Make volumes if set to true
MakeVolumes: true
# Max amount of volumes
MaxVolumes: 150
# Minimum volume size in bytes (will be a adjusted to be nearest multiple of ArticleSize)
VolumeSize: 153600000
# Absolute path to rar exe (https://www.rarlab.com/download.htm)
RarExe: "C:\\Tools\\Rar\\rar.exe"

# Par settings
# Make par2 files if set to true
MakePar2: true
# Redundancy in %
Redundancy: 10
# Absolute path to par exe
# either par2(.exe) (https://github.com/animetosho/par2cmdline-turbo/releases)
# or parpar(.exe) (https://github.com/animetosho/ParPar/releases)
Par2Exe: "C:\\Tools\\Par2\\par2.exe"

# Path settings
# All paths must be absolute paths or are treated as relative paths to the user's home folder
# Temporary path for rar/par2 files (if left empty default temp path is used)
TempPath: "D:/upper/tmp"
# Path for the NZB file
NzbPath: "D:/upper/nzbs"
# Path for the logfile (leave empty to disable logging)
LogFilePath: "D:/upper/logs"

# Miscellaneous settings
# CSV file path to log title, header, password, groups and date (leave empty to disable)
CsvPath: "D:/upper/postinglist.csv"
# CSV file delimiter
CsvDelimiter: ";"
# Delete temp folder at program end
DelTempFolder: true
# Delete input folder after successful upload
DelInputFolder: false

# Verbosity level of cmd output
# 0 = no output except for fatal errors
# 1 = outputs information
# 2 = outputs information and non-fatal errors (warnings)
# 3 = outputs information, non-fatal errors (warnings) and additional debug information if activated
Verbose: 2

# Debug mode (logs additional debug information)
Debug: true
`
}
