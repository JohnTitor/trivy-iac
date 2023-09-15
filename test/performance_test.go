package test

import (
	"context"
	"fmt"
	"io/fs"
	"testing"

	"github.com/nikpivkin/trivy-iac/pkg/scanners/terraform/executor"
	"github.com/nikpivkin/trivy-iac/pkg/scanners/terraform/parser"
	"github.com/nikpivkin/trivy-iac/test/testutil"
	"github.com/simar7/trivy-misconf-rules/pkg/rules"
)

func BenchmarkCalculate(b *testing.B) {

	f, err := createBadBlocks()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.New(f, "", parser.OptionStopOnHCLError(true))
		if err := p.ParseFS(context.TODO(), "project"); err != nil {
			b.Fatal(err)
		}
		modules, _, err := p.EvaluateAll(context.TODO())
		if err != nil {
			b.Fatal(err)
		}
		_, _, _ = executor.New().Execute(modules)
	}
}

func createBadBlocks() (fs.FS, error) {

	files := make(map[string]string)

	files["/project/main.tf"] = `
module "something" {
	source = "../modules/problem"
}
`

	for _, rule := range rules.GetRegistered() {
		if rule.Rule().Terraform == nil {
			continue
		}
		for i, bad := range rule.Rule().Terraform.BadExamples {
			filename := fmt.Sprintf("/modules/problem/%s-%d.tf", rule.Rule().LongID(), i)
			files[filename] = bad
		}
	}

	f := testutil.CreateFS(&testing.T{}, files)
	return f, nil
}