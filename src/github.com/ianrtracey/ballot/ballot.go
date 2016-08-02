package ballot

type Ballot struct {
	Topic string
	Items []string
	Votes map[string]string
}

func (ballot *Ballot) AddItem(item string) []string {
	ballot.Items = append(ballot.Items, item)
	return ballot.Items
}
