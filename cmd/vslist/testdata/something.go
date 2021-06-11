package testdata

// Something ...
//go:generate vslist -type=Something -debug
type Something struct {
	id string
}

func (s Something) String() string {
	return "testdata." + s.id
}

var (
	SomethingOne = Something{
		id: "one",
	}
	SomethingTwo = Something{
		id: "two",
	}
)

var (
	SomethingThree Something = Something{
		id: "three",
	}
)
