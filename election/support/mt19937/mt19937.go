// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package mt19937

const (
	n         = 312
	m         = 156
	notSeeded = n + 1

	hiMask uint64 = 0xffffffff80000000
	loMask uint64 = 0x000000007fffffff

	matrixA uint64 = 0xB5026F5AA96619E9
)

// MT19937 is the structure to hold the state of one instance of the
// Mersenne Twister PRNG.  New instances can be allocated using the
// mt19937.New() function.  MT19937 implements the rand.Source
// interface and rand.New() from the math/rand package can be used to
// generate different distributions from a MT19937 PRNG.
//
// This class is not safe for concurrent accesss by different
// goroutines.  If more than one goroutine accesses the PRNG, the
// callers must synchronise access using sync.Mutex or similar.
type MT19937 struct {
	state []uint64
	index int
}

// New allocates a new instance of the 64bit Mersenne Twister.
// A seed can be set using the .Seed() or .SeedFromSlice() methods.
// If no seed is set explicitly, a default seed is used instead.
func New() *MT19937 {
	res := &MT19937{
		state: make([]uint64, n),
		index: notSeeded,
	}
	return res
}

// Seed uses the given 64bit value to initialise the generator state.
// This method is part of the rand.Source interface.
func (mt *MT19937) Seed(seed int64) {
	x := mt.state
	x[0] = uint64(seed)
	for i := uint64(1); i < n; i++ {
		x[i] = 6364136223846793005*(x[i-1]^(x[i-1]>>62)) + i
	}
	mt.index = n
}

// SeedFromSlice uses the given slice of 64bit values to set the
// generator state.
func (mt *MT19937) SeedFromSlice(key []uint64) {
	mt.Seed(19650218)

	x := mt.state
	i := uint64(1)
	j := 0
	k := len(key)
	if n > k {
		k = n
	}
	for k > 0 {
		x[i] = (x[i] ^ ((x[i-1] ^ (x[i-1] >> 62)) * 3935559000370003845) +
			key[j] + uint64(j))
		i++
		if i >= n {
			x[0] = x[n-1]
			i = 1
		}
		j++
		if j >= len(key) {
			j = 0
		}
		k--
	}
	for j := uint64(0); j < n-1; j++ {
		x[i] = x[i] ^ ((x[i-1] ^ (x[i-1] >> 62)) * 2862933555777941757) - i
		i++
		if i >= n {
			x[0] = x[n-1]
			i = 1
		}
	}
	x[0] = 1 << 63
}

// Uint64 generates a (pseudo-)random 64bit value.  The output can be
// used as a replacement for a sequence of independent, uniformly
// distributed samples in the range 0, 1, ..., 2^64-1.  This method is
// part of the rand.Source64 interface.
func (mt *MT19937) Uint64() uint64 {
	x := mt.state
	if mt.index >= n {
		if mt.index == notSeeded {
			mt.Seed(5489) // default seed, as in mt19937-64.c
		}
		for i := 0; i < n-m; i++ {
			y := (x[i] & hiMask) | (x[i+1] & loMask)
			x[i] = x[i+m] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		for i := n - m; i < n-1; i++ {
			y := (x[i] & hiMask) | (x[i+1] & loMask)
			x[i] = x[i+(m-n)] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		y := (x[n-1] & hiMask) | (x[0] & loMask)
		x[n-1] = x[m-1] ^ (y >> 1) ^ ((y & 1) * matrixA)
		mt.index = 0
	}
	y := x[mt.index]
	y ^= (y >> 29) & 0x5555555555555555
	y ^= (y << 17) & 0x71D67FFFEDA60000
	y ^= (y << 37) & 0xFFF7EEE000000000
	y ^= (y >> 43)
	mt.index++
	return y
}

// Int63 generates a (pseudo-)random 63bit value.  The output can be
// used as a replacement for a sequence of independent, uniformly
// distributed samples in the range 0, 1, ..., 2^63-1.  This method is
// part of the rand.Source interface.
func (mt *MT19937) Int63() int64 {
	x := mt.state
	if mt.index >= n {
		if mt.index == notSeeded {
			mt.Seed(5489) // default seed, as in mt19937-64.c
		}
		for i := 0; i < n-m; i++ {
			y := (x[i] & hiMask) | (x[i+1] & loMask)
			x[i] = x[i+m] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		for i := n - m; i < n-1; i++ {
			y := (x[i] & hiMask) | (x[i+1] & loMask)
			x[i] = x[i+(m-n)] ^ (y >> 1) ^ ((y & 1) * matrixA)
		}
		y := (x[n-1] & hiMask) | (x[0] & loMask)
		x[n-1] = x[m-1] ^ (y >> 1) ^ ((y & 1) * matrixA)
		mt.index = 0
	}
	y := x[mt.index]
	y ^= (y >> 29) & 0x5555555555555555
	y ^= (y << 17) & 0x71D67FFFEDA60000
	y ^= (y << 37) & 0xFFF7EEE000000000
	y ^= (y >> 43)
	mt.index++
	return int64(y & 0x7fffffffffffffff)
}

// Read fills `p` with (pseudo-)random bytes.  This method implements
// the io.Reader interface.  The returned length `n` always equals
// `len(p)` and `err` is always nil.
func (mt *MT19937) Read(p []byte) (n int, err error) {
	n = len(p)
	for len(p) >= 8 {
		val := mt.Uint64()
		p[0] = byte(val)
		p[1] = byte(val >> 8)
		p[2] = byte(val >> 16)
		p[3] = byte(val >> 24)
		p[4] = byte(val >> 32)
		p[5] = byte(val >> 40)
		p[6] = byte(val >> 48)
		p[7] = byte(val >> 56)
		p = p[8:]
	}
	if len(p) > 0 {
		val := mt.Uint64()
		for i := 0; i < len(p); i++ {
			p[i] = byte(val)
			val >>= 8
		}
	}
	return n, nil
}

//////////////////////////////
type RandUniform struct {
	seed int64
	mt   [624]int64
}

func _int32(x int64) int64 {
	return int64(0xFFFFFFFF & x)
}

func RandUniformInit(seed int64) *RandUniform {
	var RanUni RandUniform
	RanUni.seed = seed
	RanUni.mt[0] = seed
	for i := 1; i < 624; i++ {
		RanUni.mt[i] = _int32(1812433253*(RanUni.mt[i-1]^RanUni.mt[i-1]>>30) + int64(i))
	}
	return &RanUni
}

func (Ruf *RandUniform) extract_number() int64 {

	Ruf.twist()
	y := Ruf.mt[0]
	y = y ^ y>>11
	y = y ^ y<<7&2636928640
	y = y ^ y<<15&4022730752
	y = y ^ y>>18
	return _int32(y)
}

func (Ruf *RandUniform) twist() {

	for i := 0; i < 624; i++ {
		y := _int32((Ruf.mt[i] & 0x80000000) + (Ruf.mt[(i+1)%624] & 0x7fffffff))
		Ruf.mt[i] = y ^ Ruf.mt[(i+397)%624]>>1
		if y%2 != 0 {
			Ruf.mt[i] = Ruf.mt[i] ^ 0x9908b0df
		}
	}
}
func powerf(x float64, n int) float64 {
	ans := 1.0

	for n != 0 {
		if n%2 == 1 {
			ans *= x
		}
		x *= x
		n /= 2
	}
	return ans
}

func (Ruf *RandUniform) Uniform(low, high float64) float64 {
	pow := powerf(2, 32)
	tmp := float64(Ruf.extract_number()) / pow
	return (high - low) * tmp
}
