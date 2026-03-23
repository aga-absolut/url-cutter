package main

import (
	"github.com/aga-absolut/url-cutter/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	)

	for _, a := range staticcheck.Analyzers {
		if a.Analyzer.Name[0] == 'S' && a.Analyzer.Name[1] == 'A' {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	analyzers = append(analyzers, analyzer.Analyzer)
	multichecker.Main(analyzers...)
}
