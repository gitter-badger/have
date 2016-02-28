package have

import (
	"fmt"
	"strings"
	"testing"
)

type generatorTestCase struct {
	source, reference string
}

func inMemTranspile(code string) (string, error) {
	parser := NewParser(NewLexer([]rune(code)))
	result, err := parser.Parse()
	if err != nil {
		return "", err
	}

	for _, stmt := range result {
		typedStmt := stmt.(ExprToProcess)
		if err := typedStmt.NegotiateTypes(); err != nil {
			return "", err
		}
	}

	cc := &CodeChunk{}

	for _, stmt := range result {
		stmt.(Generable).Generate(cc)
	}

	return cc.ReadAll(), nil
}

func testCases(t *testing.T, cases []generatorTestCase) {
	for i, c := range cases {
		result, err := inMemTranspile(c.source)
		if err != nil {
			t.Fail()
			fmt.Printf("Error compiling case %d\nError: %s\nCode: %s\n", i, err, c.source)
		}
		if a, b := strings.TrimSpace(result), strings.TrimSpace(c.reference); a != b {
			t.Fail()
			fmt.Printf("Different input and output for case %d.\nInput: `%s`\nOutput: `%s`\nWanted: `%s`\n",
				i, c.source, a, b)
		}
	}
}

func TestGenerateExpr(t *testing.T) {
	cases := []generatorTestCase{
		{source: "1", reference: "1\n"},
		{source: "\"bla\"", reference: "\"bla\"\n"},
		{source: "true", reference: "true\n"},
		{source: "false", reference: "false\n"},
		{source: "1+1", reference: "(1 + 1)\n"},
		{source: "1+(-1)", reference: "(1 + (-1))\n"},
		{source: "func a():\n 1", reference: "func a() {\n\t1\n}\n"},
		{source: "print(\"test\")", reference: "print(\"test\")\n"},
		{source: "if 1 == 2:\n 1", reference: `
if (1 == 2) {
	1
}`,
		},
		{source: "if t = 1; t == 2:\n 1", reference: `
if t := (int)(1); (t == 2) {
	1
}`,
		},
		{source: "if t = 1, k = \"aaa\"; t == 2 && k == \"bbb\":\n 1", reference: `
if t, k := (int)(1), (string)("aaa"); ((t == 2) && (k == "bbb")) {
	1
} `,
		},
		{source: "if 1 == 2:\n 1\nelse:\n 2\n", reference: `
if (1 == 2) {
	1
} else {
	2
}`},
		{source: "if 1 == 2:\n 1\nelif true == false:\n 5\nelse:\n 2\n", reference: `
if (1 == 2) {
	1
} else if (true == false) {
	5
} else {
	2
}`},
		{source: `for x = 0; x < 100; print("a"):
	print("b")`, reference: `for x := (int)(0); (x < 100); print("a") {
	print("b")
}`},
	}
	testCases(t, cases)
}