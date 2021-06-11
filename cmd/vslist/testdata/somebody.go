package testdata

// Somebody ...
//go:generate vslist -type=Somebody -debug
type Somebody struct {
	id string
}

func (s Somebody) String() string {
	return "testdata." + s.id
}

var SomebodyLiLei = Somebody{
	id: "lilei",
}

var SomebodyHanMeimei = Somebody{
	id: "hanmeimei",
}
