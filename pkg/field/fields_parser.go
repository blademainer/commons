package field

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type (
	// Parser is the default implementation of the Binder interface.
	Parser struct {
		Tag    string
		Escape bool
		Quoted bool
		// Delimiter between groups. For example: we should convert struct to http form, so this GroupDelimiter is '&' and PairDelimiter is '='
		GroupDelimiter byte
		// Delimiter between key and value.
		PairDelimiter byte
		// Field should sort by field name
		Sort bool
		// Ignore fields that value is nil
		IgnoreNilValueField bool
		fieldCache          *fieldCache
		// map[reflect.Type]encoderFunc
		encoderCache *sync.Map
	}

	fieldCache struct {
		value atomic.Value // map[reflect.Type][]field
		mu    sync.Mutex   // used only by writers
	}

	// Unmarshaler is the interface used to wrap the UnmarshalParam method.
	Unmarshaler interface {
		// UnmarshalParam decodes and assigns a value from an form or query param.
		Unmarshal(param string) error
	}

	Marshaler interface {
		Marshal() (string, error)
	}
)

func (b *Parser) String() string {
	return fmt.Sprintf("Tag: \"%s\", Escape: %v, Quoted: %v, GroupDelimiter: \"%s\", PairDelimiter: \"%s\", Sort: %v, IgnoreNilValueField: %v \n",
		b.Tag, b.Escape, b.Quoted, string(b.GroupDelimiter), string(b.PairDelimiter), b.Sort, b.IgnoreNilValueField)
}

var (
	HTTP_ENCODED_FORM_PARSER = &Parser{
		Tag:                 "form",
		Escape:              true,
		GroupDelimiter:      '&',
		PairDelimiter:       '=',
		Sort:                false,
		IgnoreNilValueField: true}
	HTTP_FORM_PARSER = &Parser{
		Tag:                 "form",
		Escape:              false,
		GroupDelimiter:      '&',
		PairDelimiter:       '=',
		Sort:                true,
		IgnoreNilValueField: true,
	}
)
