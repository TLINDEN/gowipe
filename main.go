/*
Copyright © 2022 Thomas von Dein

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/JojiiOfficial/shred"
	flag "github.com/spf13/pflag"
)

const VERSION string = "0.0.2"
const Usage string = `This is gowipe - destruct files in a non-recoverable way.

Usage: gowipe [-rcvz] <file|directory>...

Options:
-r --recursive    Delete <dir> recursively
-c --count <num>  Overwrite files <num> times
-m --mode <mode>  Use <mode> for overwriting (or use -E, -S, -M, -Z)
-n --nodelete     Do not delete files after overwriting
-N --norename     Do not rename the files
-v --verbose      Verbose output
-V --version      Show program version
-h --help         Show usage

Available modes:
zero      Overwrite with zeroes (-Z)
math      Overwrite with math random bytes (-M)
secure    Overwrite with secure random bytes (default) (-S)
encrypt   Overwrite with ChaCha2Poly1305 encryption (most secure) (-E)`

type Conf struct {
	mode     string
	count    int
	recurse  bool
	nodelete bool
	norename bool
	verbose  bool
}

func main() {
	showversion := false
	showhelp := false
	optzero := false
	optsecure := false
	optmath := false
	optencrypt := false

	c := Conf{
		verbose:  false,
		mode:     `secure`,
		count:    30,
		recurse:  false,
		nodelete: false,
		norename: false,
	}

	flag.BoolVarP(&showversion, "version", "V", showversion, "show version")
	flag.BoolVarP(&showhelp, "help", "h", showversion, "show help")
	flag.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "verbose")

	flag.StringVarP(&c.mode, "mode", "m", c.mode, "overwrite mode")

	flag.BoolVarP(&optzero, "zero", "Z", optzero, "zero mode")
	flag.BoolVarP(&optsecure, "secure", "S", optsecure, "secure mode")
	flag.BoolVarP(&optmath, "math", "M", optmath, "math mode")
	flag.BoolVarP(&optmath, "encrypt", "E", optmath, "encrypt mode")

	flag.BoolVarP(&c.recurse, "recursive", "r", c.recurse, "recursive")
	flag.BoolVarP(&c.nodelete, "nodelete", "n", c.nodelete, "don't delete")
	flag.BoolVarP(&c.norename, "norename", "N", c.norename, "don't rename")
	flag.IntVarP(&c.count, "count", "c", c.count, "overwrite count")

	flag.Parse()

	if showversion {
		fmt.Printf("This is gowipe version %s\n", VERSION)
		os.Exit(0)
	}

	if showhelp {
		fmt.Println(Usage)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println(Usage)
		os.Exit(0)
	}

	var option shred.WriteOptions

	if optzero {
		option = shred.WriteZeros
	}
	if optmath {
		option = shred.WriteRand
	}
	if optsecure {
		option = shred.WriteRandSecure
	}
	if optencrypt {
		c.mode = "encrypt"
	}

	switch c.mode {
	case `secure`:
		option = shred.WriteRandSecure
	case `math`:
		option = shred.WriteRand
	case `zero`:
		option = shred.WriteZeros
	case `encrypt`:
		optencrypt = true
	default:
		option = shred.WriteRandSecure
	}

	shredder := shred.Shredder{}
	shredconf := shred.NewShredderConf(&shredder, option, c.count, !c.nodelete)

	for _, file := range flag.Args() {
		Wipe(file, &c, shredconf)
	}
}

func Wipe(file string, c *Conf, wiper *shred.ShredderConf) {
	if info, err := os.Stat(file); err == nil {

		if info.IsDir() {
			if !c.recurse {
				fmt.Printf("-r not set, ignoring directory %s\n", file)
				return
			}

			files, err := ioutil.ReadDir(file)
			if err != nil {
				log.Fatal(err)
			}

			for _, entry := range files {
				Wipe(filepath.Join(file, entry.Name()), c, wiper)
			}

			if !c.nodelete {
				err = os.Remove(Rename(file, c))
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			if c.mode == "encrypt" {
				err := Encrypt(c, file)
				if err != nil {
					log.Fatal(err)
				}

				Rename(file, c)
			} else {
				wiper.ShredFile(Rename(file, c))
			}
		}

		if c.verbose {
			fmt.Printf("Wiped %d times: %s\n", c.count, file)
		}
	} else {
		if os.IsNotExist(err) {
			fmt.Printf("No such file or directory: %s\n", file)
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}

func Rename(file string, c *Conf) string {
	var newname string
	dir := filepath.Dir(file)
	base := filepath.Base(file)
	length := len(base)

	for i := 0; i < c.count; i++ {
		for {
			switch c.mode {
			case `secure`:
				new, err := GenerateSecureRandomString(length)
				if err != nil {
					log.Fatal(err)
				}
				newname = new
			case `math`:
				newname = GenerateMathRandomString(length)
			case `zero`:
				newname = strings.Repeat("0", length)
			}
			if newname != base {
				break
			}
		}

		/*
			 	if c.verbose {
				  fmt.Printf("renaming %s/%s => %s/%s\n", dir, base, dir, newname)
			    }
		*/

		err := os.Rename(filepath.Join(dir, base), filepath.Join(dir, newname))
		if err != nil {
			log.Fatal(err)
		}

		base = newname
	}

	return filepath.Join(dir, newname)
}
