package crypto

import (
	"encoding/binary"
	"time"
)

type Generator struct {
	seed uint64
}

func NewGenerator() *Generator {
	return &Generator{seed: uint64(time.Now().UnixNano())}
}

func (g *Generator) next() uint64 {
	g.seed = g.seed ^ (g.seed << 13)
	g.seed = g.seed ^ (g.seed >> 7)
	g.seed = g.seed ^ (g.seed << 17)
	return g.seed
}

func (g *Generator) GenerateBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i += 8 {
		val := g.next()
		remaining := length - i
		if remaining >= 8 {
			binary.LittleEndian.PutUint64(bytes[i:], val)
		} else {
			temp := make([]byte, 8)
			binary.LittleEndian.PutUint64(temp, val)
			copy(bytes[i:], temp[:remaining])
		}
	}
	return bytes
}

func (g *Generator) GenerateString(length int, charset string) string {
	bytes := g.GenerateBytes(length)
	result := make([]byte, length)
	charsetLen := len(charset)

	for i := 0; i < length; i++ {
		result[i] = charset[bytes[i]%byte(charsetLen)]
	}

	return string(result)
}
