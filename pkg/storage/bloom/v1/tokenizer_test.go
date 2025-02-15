package v1

import (
	"bufio"
	"encoding/binary"
	"os"
	"testing"

	"github.com/grafana/loki/pkg/logproto"

	"github.com/stretchr/testify/require"
)

const BigFile = "../../../logql/sketch/testdata/war_peace.txt"

var (
	twoSkipOne = NewNGramTokenizer(2, 3, 1)
	three      = NewNGramTokenizer(3, 4, 0)
	threeSkip1 = NewNGramTokenizer(3, 4, 1)
	threeSkip2 = NewNGramTokenizer(3, 4, 2)
	four       = NewNGramTokenizer(4, 5, 0)
	fourSkip1  = NewNGramTokenizer(4, 5, 1)
	fourSkip2  = NewNGramTokenizer(4, 5, 2)
	five       = NewNGramTokenizer(5, 6, 0)
	six        = NewNGramTokenizer(6, 7, 0)
)

func TestNGramIterator(t *testing.T) {
	var (
		three      = NewNGramTokenizerV2(3, 0)
		threeSkip1 = NewNGramTokenizerV2(3, 1)
		threeSkip3 = NewNGramTokenizerV2(3, 3)
	)

	for _, tc := range []struct {
		desc  string
		t     *NGramTokenizerV2
		input string
		exp   []string
	}{
		{
			t:     three,
			input: "abcdefg",
			exp:   []string{"abc", "bcd", "cde", "def", "efg"},
		},
		{
			t:     threeSkip1,
			input: "abcdefg",
			exp:   []string{"abc", "cde", "efg"},
		},
		{
			t:     threeSkip3,
			input: "abcdefgh",
			exp:   []string{"abc", "efg"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			itr := tc.t.Tokens(tc.input)
			for _, exp := range tc.exp {
				require.True(t, itr.Next())
				require.Equal(t, exp, string(itr.At()))
			}
			require.False(t, itr.Next())
		})
	}
}

func TestNGrams(t *testing.T) {
	tokenizer := NewNGramTokenizer(2, 4, 0)
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "two chars",
			input: "ab",
			exp:   []Token{{Key: []byte("ab")}},
		},
		{
			desc:  "three chars",
			input: "abc",
			exp:   []Token{{Key: []byte("ab")}, {Key: []byte("bc")}, {Key: []byte("abc")}},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("ab")}, {Key: []byte("bc")}, {Key: []byte("abc")}, {Key: []byte("cd")}, {Key: []byte("bcd")}},
		},
		{
			desc:  "foo",
			input: "日本語",
			exp:   []Token{{Key: []byte("日本")}, {Key: []byte("本語")}, {Key: []byte("日本語")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func TestNGramsSkip(t *testing.T) {

	for _, tc := range []struct {
		desc      string
		tokenizer *NgramTokenizer
		input     string
		exp       []Token
	}{
		{
			desc:      "four chars",
			tokenizer: twoSkipOne,
			input:     "abcd",
			exp:       []Token{{Key: []byte("ab")}, {Key: []byte("cd")}},
		},
		{
			desc:      "special chars",
			tokenizer: twoSkipOne,
			input:     "日本語",
			exp:       []Token{{Key: []byte("日本")}},
		},
		{
			desc:      "multi",
			tokenizer: NewNGramTokenizer(2, 4, 1),
			input:     "abcdefghij",
			exp: []Token{
				{Key: []byte("ab")},
				{Key: []byte("abc")},
				{Key: []byte("cd")},
				{Key: []byte("cde")},
				{Key: []byte("ef")},
				{Key: []byte("efg")},
				{Key: []byte("gh")},
				{Key: []byte("ghi")},
				{Key: []byte("ij")},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tc.tokenizer.Tokens(tc.input))
		})
	}
}

func Test3GramSkip0Tokenizer(t *testing.T) {
	tokenizer := three
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{{Key: []byte("abc")}},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abc")}, {Key: []byte("bcd")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test3GramSkip1Tokenizer(t *testing.T) {
	tokenizer := threeSkip1
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{{Key: []byte("abc")}},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abc")}},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{{Key: []byte("abc")}, {Key: []byte("cde")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test3GramSkip2Tokenizer(t *testing.T) {
	tokenizer := threeSkip2
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abc")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test4GramSkip0Tokenizer(t *testing.T) {
	tokenizer := four
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("bcde")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test4GramSkip1Tokenizer(t *testing.T) {
	tokenizer := fourSkip1
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "six chars",
			input: "abcdef",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("cdef")}},
		},
		{
			desc:  "seven chars",
			input: "abcdefg",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("cdef")}},
		},
		{
			desc:  "eight chars",
			input: "abcdefgh",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("cdef")}, {Key: []byte("efgh")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test4GramSkip2Tokenizer(t *testing.T) {
	tokenizer := fourSkip2
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "six chars",
			input: "abcdef",
			exp:   []Token{{Key: []byte("abcd")}},
		},
		{
			desc:  "seven chars",
			input: "abcdefg",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("defg")}},
		},
		{
			desc:  "eight chars",
			input: "abcdefgh",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("defg")}},
		},
		{
			desc:  "nine chars",
			input: "abcdefghi",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("defg")}},
		},
		{
			desc:  "ten chars",
			input: "abcdefghij",
			exp:   []Token{{Key: []byte("abcd")}, {Key: []byte("defg")}, {Key: []byte("ghij")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test5GramSkip0Tokenizer(t *testing.T) {
	tokenizer := five
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{{Key: []byte("abcde")}},
		},
		{
			desc:  "six chars",
			input: "abcdef",
			exp:   []Token{{Key: []byte("abcde")}, {Key: []byte("bcdef")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func Test6GramSkip0Tokenizer(t *testing.T) {
	tokenizer := six
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "three char",
			input: "abc",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp:   []Token{},
		},
		{
			desc:  "five chars",
			input: "abcde",
			exp:   []Token{},
		},
		{
			desc:  "six chars",
			input: "abcdef",
			exp:   []Token{{Key: []byte("abcdef")}},
		},
		{
			desc:  "seven chars",
			input: "abcdefg",
			exp:   []Token{{Key: []byte("abcdef")}, {Key: []byte("bcdefg")}},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tokenizer.Tokens(tc.input))
		})
	}
}

func makeBuf(from, through, checksum int) []byte {
	p := make([]byte, 0, 256)
	i64buf := make([]byte, binary.MaxVarintLen64)
	i32buf := make([]byte, 4)

	binary.PutVarint(i64buf, int64(from))
	p = append(p, i64buf...)
	binary.PutVarint(i64buf, int64(through))
	p = append(p, i64buf...)
	binary.LittleEndian.PutUint32(i32buf, uint32(checksum))
	p = append(p, i32buf...)
	return p
}

func TestWrappedTokenizer(t *testing.T) {
	tokenizer := threeSkip2
	for _, tc := range []struct {
		desc  string
		input string
		exp   []Token
	}{
		{
			desc:  "empty",
			input: "",
			exp:   []Token{},
		},
		{
			desc:  "single char",
			input: "a",
			exp:   []Token{},
		},
		{
			desc:  "four chars",
			input: "abcd",
			exp: []Token{
				{Key: append(makeBuf(0, 999999, 1), []byte("abc")...)},
				{Key: []byte("abc")}},
		},
		{
			desc:  "uuid",
			input: "2b1a5e46-36a2-4694-a4b1-f34cc7bdfc45",
			exp: []Token{
				{Key: append(makeBuf(0, 999999, 1), []byte("2b1")...)},
				{Key: []byte("2b1")},
				{Key: append(makeBuf(0, 999999, 1), []byte("a5e")...)},
				{Key: []byte("a5e")},
				{Key: append(makeBuf(0, 999999, 1), []byte("46-")...)},
				{Key: []byte("46-")},
				{Key: append(makeBuf(0, 999999, 1), []byte("36a")...)},
				{Key: []byte("36a")},
				{Key: append(makeBuf(0, 999999, 1), []byte("2-4")...)},
				{Key: []byte("2-4")},
				{Key: append(makeBuf(0, 999999, 1), []byte("694")...)},
				{Key: []byte("694")},
				{Key: append(makeBuf(0, 999999, 1), []byte("-a4")...)},
				{Key: []byte("-a4")},
				{Key: append(makeBuf(0, 999999, 1), []byte("b1-")...)},
				{Key: []byte("b1-")},
				{Key: append(makeBuf(0, 999999, 1), []byte("f34")...)},
				{Key: []byte("f34")},
				{Key: append(makeBuf(0, 999999, 1), []byte("cc7")...)},
				{Key: []byte("cc7")},
				{Key: append(makeBuf(0, 999999, 1), []byte("bdf")...)},
				{Key: []byte("bdf")},
				{Key: append(makeBuf(0, 999999, 1), []byte("c45")...)},
				{Key: []byte("c45")},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			chunkTokenizer := ChunkIDTokenizer(tokenizer)
			chunkTokenizer.Reinit(logproto.ChunkRef{From: 0, Through: 999999, Checksum: 1})
			require.Equal(t, tc.exp, chunkTokenizer.Tokens(tc.input))
		})
	}
}

const lorem = `
lorum ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat
duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur
sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est
laborum ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat
duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur
sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est
`

func BenchmarkTokens(b *testing.B) {
	var (
		v2Three      = NewNGramTokenizerV2(3, 0)
		v2ThreeSkip1 = NewNGramTokenizerV2(3, 1)

		// fp + from + through + checksum
		chunkPrefixLen = 8 + 8 + 8 + 4
	)

	type impl struct {
		desc string
		f    func()
	}
	type tc struct {
		desc  string
		impls []impl
	}
	for _, tc := range []tc{
		{
			desc: "three",
			impls: []impl{
				{
					desc: "v1",
					f: func() {
						for _, tok := range three.Tokens(lorem) {
							_ = tok
						}
					},
				},
				{
					desc: "v2",
					f: func() {
						itr := v2Three.Tokens(lorem)
						for itr.Next() {
							_ = itr.At()
						}
					},
				},
			},
		},
		{
			desc: "threeSkip1",
			impls: []impl{
				{
					desc: "v1",
					f: func() {
						for _, tok := range threeSkip1.Tokens(lorem) {
							_ = tok
						}
					},
				},
				{
					desc: "v2",
					f: func() {
						itr := v2ThreeSkip1.Tokens(lorem)
						for itr.Next() {
							_ = itr.At()
						}
					},
				},
			},
		},
		{
			desc: "threeChunk",
			impls: []impl{
				{
					desc: "v1",
					f: func() func() {
						chunkTokenizer := ChunkIDTokenizer(three)
						chunkTokenizer.Reinit(logproto.ChunkRef{})
						return func() {
							for _, tok := range chunkTokenizer.Tokens(lorem) {
								_ = tok
							}
						}
					}(),
				},
				{
					desc: "v2",
					f: func() func() {
						prefix := make([]byte, chunkPrefixLen, 512)
						return func() {
							itr := NewPrefixedTokenIter(prefix, v2Three.Tokens(lorem))
							for itr.Next() {
								_ = itr.At()
							}
						}
					}(),
				},
			},
		},
		{
			desc: "threeSkip1Chunk",
			impls: []impl{
				{
					desc: "v1",
					f: func() func() {
						chunkTokenizer := ChunkIDTokenizer(threeSkip1)
						chunkTokenizer.Reinit(logproto.ChunkRef{})
						return func() {
							for _, tok := range chunkTokenizer.Tokens(lorem) {
								_ = tok
							}
						}
					}(),
				},
				{
					desc: "v2",
					f: func() func() {
						prefix := make([]byte, chunkPrefixLen, 512)
						return func() {
							itr := NewPrefixedTokenIter(prefix, v2ThreeSkip1.Tokens(lorem))
							for itr.Next() {
								_ = itr.At()
							}
						}
					}(),
				},
			},
		},
	} {
		b.Run(tc.desc, func(b *testing.B) {
			for _, impl := range tc.impls {
				b.Run(impl.desc, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						impl.f()
					}
				})
			}
		})
	}
}

func BenchmarkWrappedTokens(b *testing.B) {
	chunkTokenizer := ChunkIDTokenizer(three)
	chunkTokenizer.Reinit(logproto.ChunkRef{From: 0, Through: 999999, Checksum: 1})
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		file, _ := os.Open(BigFile)
		defer file.Close()
		scanner := bufio.NewScanner(file)

		b.StartTimer()
		for scanner.Scan() {
			line := scanner.Text()
			_ = chunkTokenizer.Tokens(line)
		}
	}
}
