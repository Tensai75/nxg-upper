[![Release Workflow](https://github.com/Tensai75/nxg-upper/actions/workflows/build_and_publish.yml/badge.svg?event=release)](https://github.com/Tensai75/nxg-upper/actions/workflows/build_and_publish.yml)
[![Latest Release)](https://img.shields.io/github/v/release/Tensai75/nxg-upper?logo=github)](https://github.com/Tensai75/nxg-upper/releases/latest)

# NxG Upper
Proof of Concept for a new way of binary upload to Usenet using the NxG Header, eliminating the need for Usenet search engines and NZB files.

Uploads made with NxG Upper are 100% compatible with the standard binary upload method and the uploads are indexed by Usenet search engines. NxG Upper also generates an NZB file that is compatible with standard download tools such as SABnzbd or NZBGet. However, the message IDs for the uploaded articles are calculated based on the NxG header, so download tools compatible with the NxG header, e.g. [NxG Loader](https://github.com/Tensai75/nxg-loader/), can download the Usenet posts without the need for Usenet search engines or NZB files, since they use only the NxG header as source.

## Advantages of the NxG Header
With the NxG Header, neither Usenet search engines nor NZB files are needed for binary downloads. The message IDs for the articles are calculated directly from the NxG Header.

## Calculation of the NxG Header
The basis for the NxG header is a 21-byte random string. The random string is used as the filename for the rar archives (if enabled) and/or the par2 files (if enabled). Before uploading, the total number of articles for the data files and the total number of articles for the par2 files are calculated and overlaid with the random string:

- Random 21-byte string: `Vpjnfhd01U288QM1zxElf`
- Total data articles: `415`
- Total par2 articles: `49`
- Resulting string: `Vpjnfhd01U288Q:415:49`

The resulting string is then base64 encoded to give the resulting 28-byte NxG Header:

NxG Header: `VnBqbmZoZDAxVTI4OFE6NDE1OjQ5`

## Calculation of the message IDs
For each article, a 64-byte SHA256 hash of the following string is calculated:

`[NXGHEADER]:[ARTICLETYPE]:[ARTICLENUMBER]`

- `[NXGHEADER]` = NxG header as computed above
- `[ARTICLETYPE]` = either "data" if it is a data article, or "par2" if it is a par2 article
- `[ARTICLENUMBER]` = consecutive number of the article, separated for "data" and "par2" articles

The SHA256 hash is then split into 3 parts to create the email-like message ID:

`SHA256[0:40] + "@" + SHA256[40:61] + "." + SHA256[61:64]`

This way, the message IDs can be easily calculated from the base64-decoded NxG Header.

## Requirements
- For creating rar archives NxG Upper requires that rar.exe is installed on your system.
- For creating par2 files NxG Upper requires that either par2.exe (par2cmdline) or parpar.exe is installed on your system.

The paths to the executables have to be specified in the nxg-upper.conf configuration file.

The required executables can be downloaded here:
- rar: https://www.rarlab.com/download.htm
- par2: https://github.com/animetosho/par2cmdline-turbo/releases
- parpar: https://github.com/animetosho/ParPar/releases

## Installation
1. Download the executable file for your system from the release page.
2. Extract the archive to a folder and run the executable.
3. An nxg-upper.conf configuration file is created in this folder (or in "~/.conf/" for Linux systems).
4. Edit the nxg-upper.conf according to your requirements.

## Running the program
Run the program in a cmd line with the following argument:

`nxg-upper "[UPLOADPATH]"`

- `[UPLOADPATH]` = Path to the folder you want to upload to the Usenet

See the other command line arguments and options with:

`nxg-upper -h`

Please also read the nxg-upper.conf for additional explanations in the comments

## Todos
A lot...

This is a Proof of Concept with the minimum necessary features. 
So there is certainly a lot left to do.

## Version history
### beta 7
- heavy refactoring and bug fixes
- implement new option for header check
- use nntpPool instead of nntp
- separate option for obfuscate poster

### beta 6
- better obfuscation incl. poster
- make rar file name obfuscation optional
- new option to obfuscate yenc header

### beta 5
- bug fix: use explicit file names for par2 [fixes https://github.com/Tensai75/nxg-upper/issues/1]
- bug fix: make LogFilePath relative to home if not absolute
- bug fix: make CsvPath relative to home if not absolute

### beta 4
- updated dependencies
- fixed build script
- new option "Obfuscate" to obfuscate the message subject
- new option "PasswordLength" to set the length of the generated password
- some refactoring and additional debug output for scripting purposes

### beta 3
- include parpar as an alternative par2 executable
- change progressbar from parts to uploaded bytes and include bitrate
- bug fix for appending groups to nzb file
- bug fix for re-adding failed articles blocking the queue

### beta 2
- reduce NxG Header to 28 bytes for better compatibility with the search engines

### beta 1
- first public version

## Credits
This software is built using golang ([License](https://go.dev/LICENSE)).

This software uses the following external libraries:
- github.com/acarl005/stripansi ([License](https://github.com/acarl005/stripansi/blob/master/LICENSE))
- github.com/alexflint/go-arg ([License](https://github.com/alexflint/go-arg/blob/master/LICENSE))
- github.com/schollz/progressbar/v3 ([License](https://github.com/schollz/progressbar/blob/main/LICENSE))
- github.com/spf13/viper ([License](https://github.com/spf13/viper/blob/master/LICENSE))
