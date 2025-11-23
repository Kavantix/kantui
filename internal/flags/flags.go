package flags

import "flag"

type Context struct {
	remigrateCount *int
	debug          *bool
	dbFolder       *string
}

func New() *Context {
	c := Context{
		remigrateCount: flag.Int("remigrate", 0, "the amount of migrations to down before running up migrations"),
		debug:          flag.Bool("debug", false, "turns on debug logging"),
		dbFolder:       flag.String("db", "", "location where the database is stored"),
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

func (c *Context) DbFolder() string {
	return *c.dbFolder
}
