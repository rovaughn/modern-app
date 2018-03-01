package main

type resolver struct{}

func (r resolver) Hello(args struct{ Name string }) string {
	return "Hello " + args.Name + "!"
}
