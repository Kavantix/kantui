package flags

import "flag"

type Context struct {
	remigrateCount *int
	debug          *bool
}

func New() *Context {
	c := Context{
		remigrateCount: flag.Int("remigrate", 0, "the amount of migrations to down before running up migrations"),
		debug:          flag.Bool("debug", false, "turns on debug logging"),
	}
	flag.Parse()
	return &c
}

func (c *Context) RemigrateCount() int {
	return max(0, *c.remigrateCount)
}

func (c *Context) Debug() bool {
	return *c.debug
}
