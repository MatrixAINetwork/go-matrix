package support

type SortStrallyint []Strallyint

func (self SortStrallyint) Len() int {
	return len(self)
}
func (self SortStrallyint) Less(i, j int) bool {
	return self[i].Value > self[j].Value
}
func (self SortStrallyint) Swap(i, j int) {
	temp := self[i]
	self[i] = self[j]
	self[j] = temp
}
