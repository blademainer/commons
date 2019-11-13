package generator

import (
	"fmt"
	"github.com/blademainer/commons/pkg/sign"
	"github.com/blademainer/commons/pkg/util"
	"hash/fnv"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Generator struct {
	ClusterId   *string
	MachineId   *string
	Concurrency int
	maxIndex    uint32
	indexWidth  int
	index       uint32
}

// yyyyMMddHHmmSS
const TIME_LAYOUT = "20060102150405"
const ZERO_BYTE = byte('0')

func New(clusterId *string, concurrency int) *Generator {
	g := &Generator{}
	id := fmt.Sprint(getIdentityId())
	g.MachineId = &id
	g.ClusterId = clusterId
	g.Concurrency = concurrency
	g.maxIndex = uint32(concurrency)
	for g.indexWidth = 0; concurrency > 0; g.indexWidth++ {
		concurrency = concurrency / 10
	}
	return g
}

func getIdentityId() uint32 {
	if name, err := os.Hostname(); err == nil {
		h := fnv.New32()
		h.Write([]byte(name))
		sum32 := h.Sum32()
		return sum32
	} else {
		macs := util.GetMacAddrs()

		mac := ""
		if len(macs) == 0 {
			rsaGenerator, err := sign.NewRsa2048Generator()
			if err != nil {
				mac = util.RandString(64)
			} else {
				mac, err = rsaGenerator.GeneratePemPublicKey()
				if err != nil {
					mac = util.RandString(64)
				}
			}
		}
		h := fnv.New32()
		h.Write([]byte(mac))
		sum32 := h.Sum32()
		return sum32
	}
}

func dateStr() string {
	date := time.Now().Format(TIME_LAYOUT)
	nanosecond := time.Now().Nanosecond() / 1000
	return fmt.Sprintf("%s%d", date, nanosecond)
}

func (g *Generator) GenerateIndex() string {
	index := atomic.AddUint32(&g.index, 1) % g.maxIndex
	s := fmt.Sprint(index)
	leastSize := g.indexWidth - len(s)

	builder := strings.Builder{}
	if leastSize > 0 {
		bytes := make([]byte, leastSize)
		for i := 0; i < leastSize; i++ {
			bytes[i] = ZERO_BYTE
		}
		builder.WriteString(string(bytes))
	}
	builder.WriteString(s)
	return builder.String()
}

func (g *Generator) GenerateId() string {
	builder := strings.Builder{}
	dateStr := dateStr()
	builder.WriteString(dateStr)
	builder.WriteString(*g.ClusterId)
	builder.WriteString(*g.MachineId)
	builder.WriteString(g.GenerateIndex())
	return builder.String()
}
