package storetest

import (
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/cucumber/godog"
)

func init() {
	godog.Format("simple", "Simple formatter", simpleFormatterFunc)
}

func simpleFormatterFunc(suite string, out io.Writer) godog.Formatter { //nolint:ireturn
	return &simpleFmt{
		BaseFmt: godog.NewBaseFmt(suite, out),
		out:     out,
	}
}

type simpleFmt struct {
	*godog.BaseFmt
	out io.Writer
}

func (f *simpleFmt) Summary() { //nolint:cyclop
	passedChecks := 0
	failedChecks := 0
	passedListObjects := 0
	failedListObjects := 0
	failingString := ""

	failedFeatures := []string{}
	passedFeatures := []string{}

	for _, failingStepResult := range f.Storage.MustGetPickleStepResultsByStatus(godog.StepFailed) {
		pickle := f.Storage.MustGetPickle(failingStepResult.PickleID)
		feature := f.Storage.MustGetFeature(pickle.Uri)

		if !slices.Contains(failedFeatures, feature.Feature.Name) {
			failedFeatures = append(failedFeatures, feature.Feature.Name)
		}

		if f.isCheckStep(failingStepResult.Def.Expr) {
			failedChecks++
			failingString += failingStepResult.Err.Error() + "\n"
		} else if f.isListObjectsStep(failingStepResult.Def.Expr) {
			failedListObjects++
			failingString += failingStepResult.Err.Error() + "\n"
		}
	}

	for _, passingStepResult := range f.Storage.MustGetPickleStepResultsByStatus(godog.StepPassed) {
		pickle := f.Storage.MustGetPickle(passingStepResult.PickleID)
		feature := f.Storage.MustGetFeature(pickle.Uri)

		if !slices.Contains(passedFeatures, feature.Feature.Name) {
			passedFeatures = append(passedFeatures, feature.Feature.Name)
		}

		if f.isCheckStep(passingStepResult.Def.Expr) {
			passedChecks++
		} else if f.isListObjectsStep(passingStepResult.Def.Expr) {
			passedListObjects++
		}
	}

	unknown := ""

	for _, unknownStepResult := range f.Storage.MustGetPickleStepResultsByStatus(godog.StepUndefined) {
		pickle := f.Storage.MustGetPickleStep(unknownStepResult.PickleStepID)
		unknown += pickle.Text
	}

	if failingString != "" {
		fmt.Fprint(f.out, failingString)
	}

	fmt.Fprint(f.out, "# Test Summary #\n")
	fmt.Fprintf(f.out, "Tests %d/%d passing\n", len(passedFeatures), len(passedFeatures)+len(failedFeatures))
	fmt.Fprintf(f.out, "Checks %d/%d passing\n", passedChecks, passedChecks+failedChecks)
	fmt.Fprintf(
		f.out,
		"ListObjects %d/%d passing\n",
		passedListObjects,
		passedListObjects+failedListObjects,
	)

	if unknown != "" {
		fmt.Fprintf(f.out, "Unknown steps\n%s", unknown)
	}
}

func (f *simpleFmt) isCheckStep(expr *regexp.Regexp) bool {
	return expr == HasRelationRegex ||
		expr == HasMultiRelationsRegex ||
		expr == DoesNotHaveRelationRegex ||
		expr == DoesNotHaveMultiRelationsRegex
}

func (f *simpleFmt) isListObjectsStep(expr *regexp.Regexp) bool {
	return expr == HasPermissionsRegex || expr == DoesNotHavePermissionsRegex
}
