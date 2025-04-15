package test

type Sub struct {
	ID int64 `orm:"id key snowflake"`
	I8 int8  `orm:"i8"`
}

type Parent struct {
	ID   int     `orm:"id key auto"`
	H1   Sub     `orm:"h1"`
	H2   []Sub   `orm:"h2"`
	R3   *Sub    `orm:"r3"`
	R4   []*Sub  `orm:"r4"`
	PR4  *[]Sub  `orm:"pr4"`
	PPR4 *[]*Sub `orm:"ppr4"`
}
