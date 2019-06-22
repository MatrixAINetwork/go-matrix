// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
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
