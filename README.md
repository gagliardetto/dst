# dst

### Decorated Syntax Tree

The `dst` package attempts to provide a work-arround for [go/ast: Free-floating comments are 
single-biggest issue when manipulating the AST](https://github.com/golang/go/issues/20744).

### Example:

```go
func Example_Decorations() {
	code := `package main

	func main() {
		var a int
		a++
		print(a)
	}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "a.go", code, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	df := decorator.Decorate(f, fset)
	df = dstutil.Apply(df, func(c *dstutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *dst.DeclStmt:
			n.Decs.End.Replace("// foo")
		case *dst.IncDecStmt:
			n.Decs.AfterX.Add("/* bar */")
		case *dst.CallExpr:
			n.Decs.AfterLparen.Add("\n")
			n.Decs.AfterArgs.Add("\n")
		}
		return true
	}, nil).(*dst.File)
	f, fset = decorator.Restore(df)
	format.Node(os.Stdout, fset, f)

	//Output:
	//package main
	//
	//func main() {
	//	var a int // foo
	//	a /* bar */ ++
	//	print(
	//		a,
	//	)
	//}
}
```

### Progress as of 2nd October

I've just finished a massive reorganisation of the code generation package. Instead of scanning the 
`go/ast` package for types, all the data needed to generate the fragger, decorator and restorer now 
resides in [gendst/fragment](https://github.com/dave/dst/blob/master/gendst/fragment/fragment.go). All 
the tests that were passing before still pass (we still have some disabled tests).

### Progress as of 16th September

Big refactor today... I'm a bit happier with the code generation. I still haven't found an 
elegant solution for the `FuncDecl` special case, but all the other special cases are fixed nicely.

The `FuncDecl` special case currently has a kludgy work around by extracting some of the generated 
code out into a separate function and re-arranging by hand. Will need manual updates every time the 
code generation is changed. Needs work. However, with this kludge all the tests pass.

Next I'm going to see if it can handle some real code by feeding it the standard library source.     

### Progress as of 15th September

[github.com/dave/dst](https://github.com/dave/dst) is a fork of the `go/ast` package with a few changes. The [decorator](https://github.com/dave/dst/tree/master/decorator) package converts from `*ast.File + *token.FileSet` to `*dst.File` and back again.

All the position fields have been removed from `dst` so it's just the location in the tree that determines the position of the tokens. Decorations (e.g. comments and newlines) are stored along with each node, and attached to the node at various points. The intention is that any place `gofmt` will allow a comment / new-line to be attached, `dst` will allow this.

I've finished a very rough prototype that works pretty well. (Take a look at [restorer_test.go](https://github.com/dave/dst/blob/master/decorator/restorer_test.go#L11) - all the tests pass apart from `FuncDecl` now).

There's several special cases that it doesn't currently handle. Right now I'm generating much of the code, so the special cases are non-trivial to implement. (e.g. Look at [FuncDecl](https://github.com/golang/go/blob/master/src/go/ast/ast.go#L927-L934) - the `func` token from the `Type` field is rendered before `Recv` and `Name`). Over the next few weeks I'll refactor and handle the special cases.

As @griesemer points out a big problem is where to attach the decorations so as you manipulate the tree they remain attached to the node you were expecting. My algorithm probably needs improvement here too (see [decorator_test.go](https://github.com/dave/dst/blob/master/decorator/decorator_test.go)), but I think it currently works well enough to be useful.

### Chat?

Feel free to create an [issue](https://github.com/dave/dst/issues) or chat in the [#dst](https://gophers.slack.com/messages/CCVL24MTQ) Gophers Slack channel.
