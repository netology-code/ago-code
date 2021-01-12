package cart

type Order struct {
	Id string
	Film *Film
	Start int64
	Duration int64
	Seats []*Seat
	Price int64
	Created int64
}

type Seat struct {
	Row uint
	Number uint
}

type Film struct {
	Id string
	Title string
}
