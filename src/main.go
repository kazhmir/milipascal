package main

import (
	"flag"
	"fmt"
	. "mpc/core"
	"mpc/pipelines"
	"mpc/testing"
	"os"
	"strconv"
	"strings"
)

var lexemes = flag.Bool("lexemes", false, "runs the lexer and prints the tokens")
var ast = flag.Bool("ast", false, "runs the lexer and parser, prints AST output")
var pir = flag.Bool("pir", false, "runs the full frontend, prints pir")
var mir = flag.Bool("mir", false, "runs the full compiler, prints mir")
var asm = flag.Bool("asm", false, "runs the full compiler, prints asm")

var test = flag.Bool("test", false, "runs tests for all files in a folder,")

var verbose = flag.Bool("v", false, "verbose tests")
var outname = flag.String("o", "", "output name of file")

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		Fatal("invalid number of arguments\n")
	}
	eval(args[0])
}

func eval(filename string) {
	checkValid()
	if !strings.Contains(filename, "/") {
		filename = "./" + filename
	}
	if *test {
		res := Test(filename)
		printResults(res)
		return
	}
	normalMode(filename)
}

func normalMode(filename string) {
	switch true {
	case *lexemes:
		lexemes, err := pipelines.Lexemes(filename)
		OkOrBurst(err)
		output := []string{}
		for _, lexeme := range lexemes {
			output = append(output, lexeme.Text)
		}
		fmt.Println(strings.Join(output, ", "))
	case *ast:
		n, err := pipelines.Ast(filename)
		OkOrBurst(err)
		fmt.Println(n)
	case *pir:
		pirP, err := pipelines.Pir(filename)
		OkOrBurst(err)
		fmt.Println(pirP)
	case *mir:
		mirP, err := pipelines.Mir(filename)
		OkOrBurst(err)
		fmt.Println(mirP)
	case *asm:
		fp, err := pipelines.Fasm(filename, *outname)
		OkOrBurst(err)
		fmt.Println(fp.Contents)
	default:
		_, err := pipelines.Compile(filename, *outname)
		OkOrBurst(err)
	}
}

func checkValid() {
	var selected = []bool{*lexemes, *ast, *pir, *mir, *asm}
	var count = 0
	for _, b := range selected {
		if b {
			count++
		}
	}
	if count > 1 {
		Fatal("only one of lex, parse, pir, mir or asm flags may be used at a time")
	}
}

func printResults(results []*testing.TestResult) {
	failed := 0
	Stdout("\n")
	for _, res := range results {
		if !res.Ok && res.Message != "" {
			Stdout(res.File + "\t" + res.Message + "\n")
		}
		if !res.Ok {
			failed += 1
		}
	}
	Stdout("\n")
	Stdout("failed: " + strconv.Itoa(failed) + "\n")
	Stdout("total: " + strconv.Itoa(len(results)) + "\n")
}

func Test(folder string) []*testing.TestResult {
	entries, err := os.ReadDir(folder)
	if err != nil {
		Fatal(err.Error() + "\n")
	}
	results := []*testing.TestResult{}
	for _, v := range entries {
		fullpath := folder + "/" + v.Name()
		if v.IsDir() {
			if *verbose {
				Stdout("\u001b[35m entering: " + fullpath + "\u001b[0m\n")
			}
			res := Test(fullpath)
			results = append(results, res...)
			if *verbose {
				Stdout("\u001b[35m leaving: " + fullpath + "\u001b[0m\n")
			}
		} else if strings.HasSuffix(v.Name(), ".mp") {
			res := testing.Test(fullpath)
			results = append(results, &res)
			if *verbose {
				Stdout("testing: " + fullpath + "\t")
				Stdout(res.String() + "\n")
			}
		}
	}
	return results
}

func OkOrBurst(e *Error) {
	if e != nil {
		Fatal(e.String() + "\n")
	}
}

func Stdout(s string) {
	os.Stdout.Write([]byte(s))
}

func Fatal(s string) {
	os.Stderr.Write([]byte(s))
	os.Exit(0)
}
