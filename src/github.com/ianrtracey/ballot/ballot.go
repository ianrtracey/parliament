package ballot

type Ballot struct {
	Topic string
	Items []string
}

func (ballot *Ballot) AddItem(item string) []string {
	ballot.Items = append(ballot.Items, item)
	return ballot.Items
}
